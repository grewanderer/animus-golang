package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/runtimeexec"
	"github.com/google/uuid"
)

type evaluationSyncerConfig struct {
	Enabled               bool
	Interval              time.Duration
	Batch                 int
	DefaultImageRef       string
	DefaultPreviewSamples int

	RunTokenSecret string
	RunTokenTTL    time.Duration
	DatapilotURL   string
}

type evaluationSyncer struct {
	logger *slog.Logger
	db     *sql.DB

	executor trainingExecutor

	enabled               bool
	interval              time.Duration
	batch                 int
	defaultImageRef       string
	defaultPreviewSamples int

	runTokenSecret string
	runTokenTTL    time.Duration
	datapilotURL   string
}

func startEvaluationSyncer(ctx context.Context, logger *slog.Logger, db *sql.DB, executor trainingExecutor, cfg evaluationSyncerConfig) {
	if db == nil || executor == nil || !cfg.Enabled {
		return
	}
	interval := cfg.Interval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	batch := cfg.Batch
	if batch <= 0 {
		batch = 25
	}
	defaultPreviewSamples := cfg.DefaultPreviewSamples
	if defaultPreviewSamples <= 0 {
		defaultPreviewSamples = 16
	}

	s := &evaluationSyncer{
		logger:                logger,
		db:                    db,
		executor:              executor,
		enabled:               cfg.Enabled,
		interval:              interval,
		batch:                 batch,
		defaultImageRef:       strings.TrimSpace(cfg.DefaultImageRef),
		defaultPreviewSamples: defaultPreviewSamples,
		runTokenSecret:        strings.TrimSpace(cfg.RunTokenSecret),
		runTokenTTL:           cfg.RunTokenTTL,
		datapilotURL:          strings.TrimSpace(cfg.DatapilotURL),
	}
	go s.run(ctx)
}

func (s *evaluationSyncer) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

func (s *evaluationSyncer) syncOnce(ctx context.Context) {
	if !s.enabled || s.db == nil || s.executor == nil {
		return
	}
	kind := strings.TrimSpace(s.executor.Kind())
	if kind == "" {
		return
	}
	if strings.TrimSpace(s.runTokenSecret) == "" || s.runTokenTTL <= 0 || strings.TrimSpace(s.datapilotURL) == "" {
		s.log("evaluation syncer disabled (missing token config)", "executor", kind)
		return
	}

	s.schedule(ctx, kind)
	s.syncExecutions(ctx, kind)
}

type evaluationCandidate struct {
	RunID            string
	DatasetVersionID string
	TrainingImageRef string
	TrainingDigest   string
	ResourcesRaw     []byte
	ParamsRaw        []byte
}

func (s *evaluationSyncer) schedule(ctx context.Context, executorKind string) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT r.run_id,
				COALESCE(r.dataset_version_id,''),
				e.image_ref,
				COALESCE(e.image_digest,''),
				e.resources,
				r.params
		 FROM experiment_runs r
		 JOIN experiment_run_executions e ON e.run_id = r.run_id
		 LEFT JOIN LATERAL (
			SELECT status, observed_at
			FROM experiment_run_state_events
			WHERE run_id = r.run_id
			ORDER BY observed_at DESC
			LIMIT 1
		 ) s ON true
		 WHERE e.executor = $1
		   AND COALESCE(s.status, r.status) = 'succeeded'
		   AND EXISTS (
			 SELECT 1
			 FROM experiment_run_artifacts a
			 WHERE a.run_id = r.run_id
			   AND a.kind = 'model'
		   )
		   AND NOT EXISTS (
			 SELECT 1
			 FROM experiment_run_evaluations ev
			 WHERE ev.run_id = r.run_id
		   )
		 ORDER BY r.started_at ASC
		 LIMIT $2`,
		executorKind,
		s.batch,
	)
	if err != nil {
		s.log("evaluation schedule query failed", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c evaluationCandidate
		if err := rows.Scan(&c.RunID, &c.DatasetVersionID, &c.TrainingImageRef, &c.TrainingDigest, &c.ResourcesRaw, &c.ParamsRaw); err != nil {
			s.log("evaluation schedule scan failed", "error", err)
			return
		}
		s.scheduleRun(ctx, executorKind, c)
	}
	if err := rows.Err(); err != nil {
		s.log("evaluation schedule rows error", "error", err)
	}
}

func (s *evaluationSyncer) scheduleRun(ctx context.Context, executorKind string, c evaluationCandidate) {
	runID := strings.TrimSpace(c.RunID)
	if runID == "" {
		return
	}
	datasetVersionID := strings.TrimSpace(c.DatasetVersionID)
	if datasetVersionID == "" {
		s.log("evaluation skipped (missing dataset_version_id)", "run_id", runID)
		return
	}

	params := map[string]any{}
	if len(c.ParamsRaw) > 0 {
		_ = json.Unmarshal(c.ParamsRaw, &params)
	}

	enabled := true
	if v, ok := params["evaluation_enabled"]; ok {
		if b, ok := v.(bool); ok {
			enabled = b
		}
	}

	previewSamples := s.defaultPreviewSamples
	if v, ok := params["eval_preview_samples"]; ok {
		switch t := v.(type) {
		case float64:
			previewSamples = int(t)
		case int:
			previewSamples = t
		case int64:
			previewSamples = int(t)
		}
	}
	previewSamples = clampInt(previewSamples, 1, 128)

	split := "evaluate"
	if raw, ok := params["eval_split"].(string); ok {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value != "" {
			split = value
		}
	}

	trainingImageRef := strings.TrimSpace(c.TrainingImageRef)
	trainingDigest := strings.TrimSpace(c.TrainingDigest)

	imageRef := trainingImageRef
	imageSource := "training_image_ref"
	if raw, ok := params["evaluation_image_ref"].(string); ok && strings.TrimSpace(raw) != "" {
		imageRef = strings.TrimSpace(raw)
		imageSource = "params:evaluation_image_ref"
	} else if s.defaultImageRef != "" {
		imageRef = s.defaultImageRef
		imageSource = "service:ANIMUS_EVALUATION_IMAGE_REF"
	}

	imageExecutionRef := imageRef
	imageDigest := ""
	var resolveErr error
	if enabled {
		if strings.EqualFold(strings.TrimSpace(executorKind), "docker") && isSHA256Digest(trainingDigest) && imageRef == trainingImageRef {
			imageExecutionRef = strings.ToLower(strings.TrimSpace(trainingDigest))
			imageDigest = imageExecutionRef
		} else {
			execRef, digest, err := resolveImageForExecutor(ctx, executorKind, s.executor, imageRef)
			if err != nil {
				resolveErr = err
			} else {
				imageExecutionRef = execRef
				imageDigest = digest
			}
		}
	}

	resources := map[string]any{}
	if len(c.ResourcesRaw) > 0 {
		_ = json.Unmarshal(c.ResourcesRaw, &resources)
	}
	if raw, ok := params["evaluation_resources"].(map[string]any); ok && len(raw) > 0 {
		resources = raw
	}
	resourcesJSON, _ := json.Marshal(resources)

	now := time.Now().UTC()
	evaluationID := uuid.NewString()

	type integrityInput struct {
		EvaluationID      string          `json:"evaluation_id"`
		RunID             string          `json:"run_id"`
		Executor          string          `json:"executor"`
		ImageRef          string          `json:"image_ref"`
		ImageDigest       string          `json:"image_digest,omitempty"`
		Resources         json.RawMessage `json:"resources"`
		DatapilotURL      string          `json:"datapilot_url"`
		PreviewSamples    int             `json:"preview_samples"`
		CreatedAt         time.Time       `json:"created_at"`
		CreatedBy         string          `json:"created_by"`
		DatasetVersionID  string          `json:"dataset_version_id"`
		ImageSource       string          `json:"image_source,omitempty"`
		EvalSplit         string          `json:"eval_split,omitempty"`
		EvaluationEnabled bool            `json:"evaluation_enabled"`
	}
	integrity, err := integritySHA256(integrityInput{
		EvaluationID:      evaluationID,
		RunID:             runID,
		Executor:          executorKind,
		ImageRef:          imageRef,
		ImageDigest:       imageDigest,
		Resources:         resourcesJSON,
		DatapilotURL:      s.datapilotURL,
		PreviewSamples:    previewSamples,
		CreatedAt:         now,
		CreatedBy:         "system",
		DatasetVersionID:  datasetVersionID,
		ImageSource:       imageSource,
		EvalSplit:         split,
		EvaluationEnabled: enabled,
	})
	if err != nil {
		s.log("evaluation integrity failed", "run_id", runID, "error", err)
		return
	}

	status := "pending"
	details := map[string]any{
		"executor":        executorKind,
		"image_ref":       imageRef,
		"image_execution": imageExecutionRef,
		"image_digest":    imageDigest,
		"image_source":    imageSource,
		"preview_samples": previewSamples,
		"eval_split":      split,
	}
	if !enabled {
		status = "canceled"
		details["reason"] = "disabled"
	} else if resolveErr != nil {
		status = "failed"
		switch {
		case errors.Is(resolveErr, runtimeexec.ErrImageRefDigestRequired):
			details["reason"] = "image_ref_digest_required"
		case errors.Is(resolveErr, runtimeexec.ErrImageRefNotFound):
			details["reason"] = "image_ref_not_found"
		default:
			details["reason"] = "image_ref_resolution_failed"
		}
		details["error"] = resolveErr.Error()
	} else if imageDigest == "" {
		status = "failed"
		details["reason"] = "image_ref_digest_required"
	}

	k8sJobName := ""
	dockerName := ""
	switch strings.TrimSpace(executorKind) {
	case "kubernetes_job":
		k8sJobName = "animus-eval-" + runID
	case "docker":
		dockerName = "animus-eval-" + runID
	default:
		status = "failed"
		details["reason"] = "unsupported_executor"
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log("evaluation begin tx failed", "run_id", runID, "error", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(
		ctx,
		`INSERT INTO experiment_run_evaluations (
			evaluation_id,
			run_id,
			executor,
			image_ref,
			image_digest,
			resources,
			k8s_namespace,
			k8s_job_name,
			docker_container_id,
			datapilot_url,
			preview_samples,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (run_id) DO NOTHING`,
		evaluationID,
		runID,
		executorKind,
		imageRef,
		nullString(imageDigest),
		resourcesJSON,
		nullString(""),
		nullString(k8sJobName),
		nullString(dockerName),
		s.datapilotURL,
		previewSamples,
		now,
		"system",
		integrity,
	)
	if err != nil {
		s.log("evaluation insert failed", "run_id", runID, "error", err)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return
	}

	inserted, err := (&experimentsAPI{db: s.db}).insertEvaluationStateEvent(ctx, tx, evaluationID, status, now, details)
	if err != nil {
		s.log("evaluation insert state failed", "run_id", runID, "status", status, "error", err)
		return
	}
	if inserted {
		level := "info"
		msg := "evaluation " + status
		if status == "failed" {
			level = "error"
		}
		if err := (&experimentsAPI{db: s.db}).insertRunEvent(ctx, tx, runID, "system", level, msg, map[string]any{"evaluation_id": evaluationID, "details": details}); err != nil {
			s.log("evaluation insert run event failed", "run_id", runID, "error", err)
			return
		}
	}

	_, err = auditlog.Insert(ctx, tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        "system",
		Action:       "experiment_run_evaluation." + status,
		ResourceType: "experiment_run_evaluation",
		ResourceID:   evaluationID,
		Payload: map[string]any{
			"service":            "experiments",
			"evaluation_id":      evaluationID,
			"run_id":             runID,
			"dataset_version_id": datasetVersionID,
			"executor":           executorKind,
			"image_ref":          imageRef,
			"image_digest":       imageDigest,
			"preview_samples":    previewSamples,
			"eval_split":         split,
			"details":            details,
		},
	})
	if err != nil {
		s.log("evaluation audit failed", "run_id", runID, "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.log("evaluation commit failed", "run_id", runID, "error", err)
		return
	}

	if status != "pending" {
		return
	}

	runToken, err := auth.GenerateRunToken(s.runTokenSecret, auth.RunTokenClaims{
		RunID:            runID,
		DatasetVersionID: datasetVersionID,
		ExpiresAtUnix:    now.Add(s.runTokenTTL).Unix(),
	}, now)
	if err != nil {
		_ = s.failEvaluation(ctx, evaluationID, runID, executorKind, imageRef, err)
		return
	}

	spec := trainingJobSpec{
		RunID:            runID,
		DatasetVersionID: datasetVersionID,
		ImageRef:         imageExecutionRef,
		DatapilotURL:     s.datapilotURL,
		Token:            runToken,
		Resources:        resources,
		K8sJobName:       k8sJobName,
		DockerName:       dockerName,
		JobKind:          "evaluation",
		Env: map[string]string{
			"ANIMUS_EVALUATION_ID":        evaluationID,
			"ANIMUS_EVAL_PREVIEW_SAMPLES": strconv.Itoa(previewSamples),
			"ANIMUS_EVAL_SPLIT":           split,
		},
	}

	if err := s.executor.Submit(ctx, spec); err != nil {
		_ = s.failEvaluation(ctx, evaluationID, runID, executorKind, imageRef, err)
		return
	}
	_ = s.markEvaluationRunning(ctx, evaluationID, runID, executorKind, imageRef)
}

type evaluationExecutionRow struct {
	EvaluationID string
	RunID        string
	Executor     string
	K8sNamespace string
	K8sJobName   string
	DockerName   string
}

func (s *evaluationSyncer) syncExecutions(ctx context.Context, executorKind string) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT e.evaluation_id,
				e.run_id,
				e.executor,
				COALESCE(e.k8s_namespace,''),
				COALESCE(e.k8s_job_name,''),
				COALESCE(e.docker_container_id,'')
		 FROM experiment_run_evaluations e
		 WHERE e.executor = $1
		   AND NOT EXISTS (
			 SELECT 1
			 FROM experiment_run_evaluation_state_events s
			 WHERE s.evaluation_id = e.evaluation_id
			   AND s.status IN ('succeeded','failed','canceled')
		   )
		 ORDER BY e.created_at ASC
		 LIMIT $2`,
		executorKind,
		s.batch,
	)
	if err != nil {
		s.log("evaluation sync query failed", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var row evaluationExecutionRow
		if err := rows.Scan(&row.EvaluationID, &row.RunID, &row.Executor, &row.K8sNamespace, &row.K8sJobName, &row.DockerName); err != nil {
			s.log("evaluation sync scan failed", "error", err)
			return
		}
		s.syncExecution(ctx, row)
	}
	if err := rows.Err(); err != nil {
		s.log("evaluation sync rows error", "error", err)
	}
}

func (s *evaluationSyncer) syncExecution(ctx context.Context, row evaluationExecutionRow) {
	obs, err := s.executor.Inspect(ctx, trainingExecution{
		RunID:           row.RunID,
		Executor:        row.Executor,
		K8sNamespace:    row.K8sNamespace,
		K8sJobName:      row.K8sJobName,
		DockerContainer: row.DockerName,
	})
	if err != nil {
		s.log("evaluation inspect failed", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "error", err)
		return
	}

	status := strings.ToLower(strings.TrimSpace(obs.Status))
	switch status {
	case "pending", "":
		return
	case "running", "succeeded", "failed":
	default:
		s.log("unexpected evaluation status", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "status", obs.Status)
		return
	}

	now := time.Now().UTC()
	details := obs.Details
	if details == nil {
		details = map[string]any{}
	}
	if obs.Message != "" {
		details["message"] = obs.Message
	}
	details["executor"] = row.Executor

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log("evaluation begin tx failed", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "error", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := (&experimentsAPI{db: s.db}).insertEvaluationStateEvent(ctx, tx, row.EvaluationID, status, now, details)
	if err != nil {
		s.log("evaluation insert state failed", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "status", status, "error", err)
		return
	}
	if !inserted {
		_ = tx.Rollback()
		return
	}

	level := "info"
	msg := "evaluation " + status
	if status == "failed" {
		level = "error"
	}
	_ = (&experimentsAPI{db: s.db}).insertRunEvent(ctx, tx, row.RunID, "system", level, msg, map[string]any{"evaluation_id": row.EvaluationID, "details": details})

	_, err = auditlog.Insert(ctx, tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        "system",
		Action:       "experiment_run_evaluation." + status,
		ResourceType: "experiment_run_evaluation",
		ResourceID:   row.EvaluationID,
		Payload: map[string]any{
			"service":       "experiments",
			"evaluation_id": row.EvaluationID,
			"run_id":        row.RunID,
			"status":        status,
			"details":       details,
		},
	})
	if err != nil {
		s.log("evaluation audit insert failed", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "status", status, "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.log("evaluation commit failed", "evaluation_id", row.EvaluationID, "run_id", row.RunID, "status", status, "error", err)
	}
}

func (s *evaluationSyncer) markEvaluationRunning(ctx context.Context, evaluationID string, runID string, executor string, imageRef string) error {
	now := time.Now().UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := (&experimentsAPI{db: s.db}).insertEvaluationStateEvent(ctx, tx, evaluationID, "running", now, map[string]any{
		"executor":        executor,
		"image_ref":       imageRef,
		"evaluation_id":   evaluationID,
		"datapilot_url":   s.datapilotURL,
		"training_run_id": runID,
	})
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}

	_ = (&experimentsAPI{db: s.db}).insertRunEvent(ctx, tx, runID, "system", "info", "evaluation running", map[string]any{"evaluation_id": evaluationID})

	_, err = auditlog.Insert(ctx, tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        "system",
		Action:       "experiment_run_evaluation.running",
		ResourceType: "experiment_run_evaluation",
		ResourceID:   evaluationID,
		Payload: map[string]any{
			"service":       "experiments",
			"evaluation_id": evaluationID,
			"run_id":        runID,
			"executor":      executor,
			"image_ref":     imageRef,
		},
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *evaluationSyncer) failEvaluation(ctx context.Context, evaluationID string, runID string, executor string, imageRef string, submitErr error) error {
	now := time.Now().UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := (&experimentsAPI{db: s.db}).insertEvaluationStateEvent(ctx, tx, evaluationID, "failed", now, map[string]any{
		"reason":        "submit_failed",
		"executor":      executor,
		"image_ref":     imageRef,
		"evaluation_id": evaluationID,
		"error":         submitErr.Error(),
	})
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}

	_ = (&experimentsAPI{db: s.db}).insertRunEvent(ctx, tx, runID, "system", "error", "evaluation failed (submit_failed)", map[string]any{"evaluation_id": evaluationID})

	_, err = auditlog.Insert(ctx, tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        "system",
		Action:       "experiment_run_evaluation.failed",
		ResourceType: "experiment_run_evaluation",
		ResourceID:   evaluationID,
		Payload: map[string]any{
			"service":       "experiments",
			"evaluation_id": evaluationID,
			"run_id":        runID,
			"executor":      executor,
			"image_ref":     imageRef,
			"error":         submitErr.Error(),
		},
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *evaluationSyncer) log(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	fields := []any{"component", "evaluation_syncer"}
	fields = append(fields, attrs...)
	if len(attrs) >= 2 {
		var err error
		for i := 0; i+1 < len(attrs); i += 2 {
			key, ok := attrs[i].(string)
			if !ok || key != "error" {
				continue
			}
			switch v := attrs[i+1].(type) {
			case error:
				err = v
			}
		}
		if err != nil && errors.Is(err, context.Canceled) {
			return
		}
	}
	s.logger.Warn(msg, fields...)
}
