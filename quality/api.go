package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/internal/platform/auth"
	"github.com/animus-labs/animus-go/internal/platform/objectstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
)

type qualityAPI struct {
	logger   *slog.Logger
	db       *sql.DB
	store    *minio.Client
	storeCfg objectstore.Config
}

func newQualityAPI(logger *slog.Logger, db *sql.DB, store *minio.Client, storeCfg objectstore.Config) *qualityAPI {
	return &qualityAPI{
		logger:   logger,
		db:       db,
		store:    store,
		storeCfg: storeCfg,
	}
}

func (api *qualityAPI) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /rules", api.handleListRules)
	mux.HandleFunc("POST /rules", api.handleCreateRule)
	mux.HandleFunc("GET /rules/{rule_id}", api.handleGetRule)

	mux.HandleFunc("POST /evaluations", api.handleCreateEvaluation)
	mux.HandleFunc("GET /evaluations", api.handleListEvaluations)
	mux.HandleFunc("GET /evaluations/{evaluation_id}", api.handleGetEvaluation)

	mux.HandleFunc("GET /gates/dataset-versions/{version_id}", api.handleGateStatus)
}

type qualityRule struct {
	RuleID      string          `json:"rule_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Spec        json.RawMessage `json:"spec"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   string          `json:"created_by"`
}

type createRuleRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Spec        json.RawMessage `json:"spec"`
}

func (api *qualityAPI) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	var req createRuleRequest
	if err := decodeJSON(r, &req); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_json")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		api.writeError(w, r, http.StatusBadRequest, "name_required")
		return
	}
	description := strings.TrimSpace(req.Description)
	if len(req.Spec) == 0 {
		api.writeError(w, r, http.StatusBadRequest, "spec_required")
		return
	}

	var spec RuleSpec
	if err := decodeJSONBytes(req.Spec, &spec); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_spec")
		return
	}
	if err := spec.Validate(); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_spec")
		return
	}

	specJSON, err := json.Marshal(spec)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().UTC()
	ruleID := uuid.NewString()

	type integrityInput struct {
		RuleID      string          `json:"rule_id"`
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		Spec        json.RawMessage `json:"spec"`
		CreatedAt   time.Time       `json:"created_at"`
		CreatedBy   string          `json:"created_by"`
	}
	integrity, err := integritySHA256(integrityInput{
		RuleID:      ruleID,
		Name:        name,
		Description: description,
		Spec:        specJSON,
		CreatedAt:   now,
		CreatedBy:   identity.Subject,
	})
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	tx, err := api.db.BeginTx(r.Context(), nil)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		r.Context(),
		`INSERT INTO quality_rules (
			rule_id,
			name,
			description,
			spec,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		ruleID,
		name,
		nullString(description),
		specJSON,
		now,
		identity.Subject,
		integrity,
	)
	if err != nil {
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "rule_name_exists")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = auditlog.Insert(r.Context(), tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        identity.Subject,
		Action:       "quality_rule.create",
		ResourceType: "quality_rule",
		ResourceID:   ruleID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":      "quality",
			"rule_id":      ruleID,
			"name":         name,
			"description":  description,
			"spec_schema":  spec.Schema,
			"checks_count": len(spec.Checks),
		},
	})
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
		return
	}

	if err := tx.Commit(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Location", "/rules/"+ruleID)
	api.writeJSON(w, http.StatusCreated, qualityRule{
		RuleID:      ruleID,
		Name:        name,
		Description: description,
		Spec:        specJSON,
		CreatedAt:   now,
		CreatedBy:   identity.Subject,
	})
}

func (api *qualityAPI) handleListRules(w http.ResponseWriter, r *http.Request) {
	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)

	rows, err := api.db.QueryContext(
		r.Context(),
		`SELECT rule_id, name, description, spec, created_at, created_by
		 FROM quality_rules
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer rows.Close()

	out := make([]qualityRule, 0, limit)
	for rows.Next() {
		var (
			ruleID      string
			name        string
			description sql.NullString
			spec        []byte
			createdAt   time.Time
			createdBy   string
		)
		if err := rows.Scan(&ruleID, &name, &description, &spec, &createdAt, &createdBy); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		out = append(out, qualityRule{
			RuleID:      ruleID,
			Name:        name,
			Description: description.String,
			Spec:        normalizeJSON(spec),
			CreatedAt:   createdAt,
			CreatedBy:   createdBy,
		})
	}
	if err := rows.Err(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"rules": out})
}

func (api *qualityAPI) handleGetRule(w http.ResponseWriter, r *http.Request) {
	ruleID := strings.TrimSpace(r.PathValue("rule_id"))
	if ruleID == "" {
		api.writeError(w, r, http.StatusBadRequest, "rule_id_required")
		return
	}

	var (
		name        string
		description sql.NullString
		spec        []byte
		createdAt   time.Time
		createdBy   string
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT name, description, spec, created_at, created_by
		 FROM quality_rules
		 WHERE rule_id = $1`,
		ruleID,
	).Scan(&name, &description, &spec, &createdAt, &createdBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, qualityRule{
		RuleID:      ruleID,
		Name:        name,
		Description: description.String,
		Spec:        normalizeJSON(spec),
		CreatedAt:   createdAt,
		CreatedBy:   createdBy,
	})
}

type createEvaluationRequest struct {
	DatasetVersionID string `json:"dataset_version_id"`
	RuleID           string `json:"rule_id,omitempty"`
}

type evaluation struct {
	EvaluationID     string          `json:"evaluation_id"`
	DatasetVersionID string          `json:"dataset_version_id"`
	RuleID           string          `json:"rule_id"`
	Status           string          `json:"status"`
	EvaluatedAt      time.Time       `json:"evaluated_at"`
	EvaluatedBy      string          `json:"evaluated_by"`
	Summary          json.RawMessage `json:"summary"`
	ReportObjectKey  string          `json:"report_object_key"`
	ReportSHA256     string          `json:"report_sha256"`
	ReportSizeBytes  int64           `json:"report_size_bytes"`
}

func (api *qualityAPI) handleCreateEvaluation(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	var req createEvaluationRequest
	if err := decodeJSON(r, &req); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_json")
		return
	}

	versionID := strings.TrimSpace(req.DatasetVersionID)
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_version_id_required")
		return
	}

	var (
		datasetID     string
		ordinal       int64
		contentSHA256 string
		objectKey     string
		metadata      []byte
		qualityRuleID sql.NullString
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT dataset_id, ordinal, content_sha256, object_key, metadata, quality_rule_id
		 FROM dataset_versions
		 WHERE version_id = $1`,
		versionID,
	).Scan(&datasetID, &ordinal, &contentSHA256, &objectKey, &metadata, &qualityRuleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	meta := normalizeJSON(metadata)
	filename := jsonFieldString(meta, "filename")
	contentType := jsonFieldString(meta, "content_type")

	ruleID := strings.TrimSpace(req.RuleID)
	if ruleID == "" {
		ruleID = strings.TrimSpace(qualityRuleID.String)
	}
	if ruleID == "" {
		api.writeError(w, r, http.StatusBadRequest, "rule_id_required")
		return
	}

	var (
		ruleName    string
		ruleSpecRaw []byte
	)
	err = api.db.QueryRowContext(
		r.Context(),
		`SELECT name, spec
		 FROM quality_rules
		 WHERE rule_id = $1`,
		ruleID,
	).Scan(&ruleName, &ruleSpecRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "rule_not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	var ruleSpec RuleSpec
	if err := decodeJSONBytes(ruleSpecRaw, &ruleSpec); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "invalid_rule_spec")
		return
	}
	if err := ruleSpec.Validate(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "invalid_rule_spec")
		return
	}

	statCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	objectStat, statErr := api.store.StatObject(statCtx, api.storeCfg.BucketDatasets, objectKey, minio.StatObjectOptions{})
	cancel()

	now := time.Now().UTC()
	evaluationID := uuid.NewString()

	dsv := datasetVersion{
		VersionID:     versionID,
		DatasetID:     datasetID,
		Ordinal:       ordinal,
		ObjectKey:     objectKey,
		ContentSHA256: contentSHA256,
		Metadata:      meta,
		QualityRuleID: strings.TrimSpace(qualityRuleID.String),
		Filename:      filename,
		ContentType:   contentType,
	}

	var report evaluationReport
	if statErr != nil {
		report = evaluationReport{
			Schema:       evaluationReportSchemaV1,
			EvaluationID: evaluationID,
			DatasetVersion: datasetVersion{
				VersionID:     dsv.VersionID,
				DatasetID:     dsv.DatasetID,
				Ordinal:       dsv.Ordinal,
				ObjectKey:     dsv.ObjectKey,
				ContentSHA256: dsv.ContentSHA256,
				SizeBytes:     0,
				ContentType:   dsv.ContentType,
				Filename:      dsv.Filename,
				Metadata:      dsv.Metadata,
				QualityRuleID: dsv.QualityRuleID,
			},
			Rule: qualityRuleInfo{
				RuleID: ruleID,
				Name:   ruleName,
			},
			Status:      "error",
			EvaluatedAt: now,
			EvaluatedBy: identity.Subject,
			Summary: summary{
				ChecksTotal: 0,
				ChecksPass:  0,
				ChecksFail:  0,
				ChecksError: 0,
			},
			Error:       "dataset object unavailable",
			RawRuleSpec: normalizeJSON(ruleSpecRaw),
		}
	} else {
		inputs := evaluationInputs{
			DatasetVersion: dsv,
			RuleID:         ruleID,
			RuleName:       ruleName,
			RuleSpec:       ruleSpec,
			RuleSpecRaw:    normalizeJSON(ruleSpecRaw),
			Object: objectInfo{
				Size:        objectStat.Size,
				ContentType: objectStat.ContentType,
			},
		}
		report = evaluate(r.Context(), now, identity.Subject, evaluationID, inputs, func(ctx context.Context, key string) (io.ReadCloser, error) {
			obj, err := api.store.GetObject(ctx, api.storeCfg.BucketDatasets, key, minio.GetObjectOptions{})
			if err != nil {
				return nil, err
			}
			if _, err := obj.Stat(); err != nil {
				_ = obj.Close()
				return nil, err
			}
			return obj, nil
		})
	}

	reportBytes, err := json.Marshal(report)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	reportObjectKey := "quality/dataset-versions/" + versionID + "/" + evaluationID + ".json"
	reportSHA := sha256.Sum256(reportBytes)
	reportSHA256 := hex.EncodeToString(reportSHA[:])

	putCtx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	_, err = api.store.PutObject(
		putCtx,
		api.storeCfg.BucketArtifacts,
		reportObjectKey,
		bytes.NewReader(reportBytes),
		int64(len(reportBytes)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	cancel()
	if err != nil {
		api.writeError(w, r, http.StatusBadGateway, "artifact_store_failed")
		return
	}

	summaryJSON, err := json.Marshal(report.Summary)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	type evalIntegrityInput struct {
		EvaluationID     string          `json:"evaluation_id"`
		DatasetVersionID string          `json:"dataset_version_id"`
		RuleID           string          `json:"rule_id"`
		Status           string          `json:"status"`
		EvaluatedAt      time.Time       `json:"evaluated_at"`
		EvaluatedBy      string          `json:"evaluated_by"`
		Summary          json.RawMessage `json:"summary"`
		ReportObjectKey  string          `json:"report_object_key"`
		ReportSHA256     string          `json:"report_sha256"`
		ReportSizeBytes  int64           `json:"report_size_bytes"`
	}
	integrity, err := integritySHA256(evalIntegrityInput{
		EvaluationID:     evaluationID,
		DatasetVersionID: versionID,
		RuleID:           ruleID,
		Status:           report.Status,
		EvaluatedAt:      now,
		EvaluatedBy:      identity.Subject,
		Summary:          summaryJSON,
		ReportObjectKey:  reportObjectKey,
		ReportSHA256:     reportSHA256,
		ReportSizeBytes:  int64(len(reportBytes)),
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	tx, err := api.db.BeginTx(r.Context(), nil)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		r.Context(),
		`INSERT INTO quality_evaluations (
			evaluation_id,
			dataset_version_id,
			rule_id,
			status,
			evaluated_at,
			evaluated_by,
			summary,
			report_object_key,
			report_sha256,
			report_size_bytes,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		evaluationID,
		versionID,
		ruleID,
		report.Status,
		now,
		identity.Subject,
		summaryJSON,
		reportObjectKey,
		reportSHA256,
		int64(len(reportBytes)),
		integrity,
	)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = auditlog.Insert(r.Context(), tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        identity.Subject,
		Action:       "quality_evaluation.create",
		ResourceType: "quality_evaluation",
		ResourceID:   evaluationID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":            "quality",
			"evaluation_id":      evaluationID,
			"dataset_version_id": versionID,
			"dataset_id":         datasetID,
			"rule_id":            ruleID,
			"status":             report.Status,
			"report_object_key":  reportObjectKey,
			"report_sha256":      reportSHA256,
		},
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
		return
	}

	if err := tx.Commit(); err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketArtifacts, reportObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Location", "/evaluations/"+evaluationID)
	api.writeJSON(w, http.StatusCreated, evaluation{
		EvaluationID:     evaluationID,
		DatasetVersionID: versionID,
		RuleID:           ruleID,
		Status:           report.Status,
		EvaluatedAt:      now,
		EvaluatedBy:      identity.Subject,
		Summary:          summaryJSON,
		ReportObjectKey:  reportObjectKey,
		ReportSHA256:     reportSHA256,
		ReportSizeBytes:  int64(len(reportBytes)),
	})
}

func (api *qualityAPI) handleListEvaluations(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.URL.Query().Get("dataset_version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_version_id_required")
		return
	}
	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)

	rows, err := api.db.QueryContext(
		r.Context(),
		`SELECT evaluation_id, rule_id, status, evaluated_at, evaluated_by, summary, report_object_key, report_sha256, report_size_bytes
		 FROM quality_evaluations
		 WHERE dataset_version_id = $1
		 ORDER BY evaluated_at DESC
		 LIMIT $2`,
		versionID,
		limit,
	)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer rows.Close()

	out := make([]evaluation, 0, limit)
	for rows.Next() {
		var (
			evaluationID    string
			ruleID          string
			status          string
			evaluatedAt     time.Time
			evaluatedBy     string
			summaryRaw      []byte
			reportObjectKey string
			reportSHA256    string
			reportSizeBytes int64
		)
		if err := rows.Scan(&evaluationID, &ruleID, &status, &evaluatedAt, &evaluatedBy, &summaryRaw, &reportObjectKey, &reportSHA256, &reportSizeBytes); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		out = append(out, evaluation{
			EvaluationID:     evaluationID,
			DatasetVersionID: versionID,
			RuleID:           ruleID,
			Status:           status,
			EvaluatedAt:      evaluatedAt,
			EvaluatedBy:      evaluatedBy,
			Summary:          normalizeJSON(summaryRaw),
			ReportObjectKey:  reportObjectKey,
			ReportSHA256:     reportSHA256,
			ReportSizeBytes:  reportSizeBytes,
		})
	}
	if err := rows.Err(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"evaluations": out})
}

func (api *qualityAPI) handleGetEvaluation(w http.ResponseWriter, r *http.Request) {
	evaluationID := strings.TrimSpace(r.PathValue("evaluation_id"))
	if evaluationID == "" {
		api.writeError(w, r, http.StatusBadRequest, "evaluation_id_required")
		return
	}

	var (
		versionID       string
		ruleID          string
		status          string
		evaluatedAt     time.Time
		evaluatedBy     string
		summaryRaw      []byte
		reportObjectKey string
		reportSHA256    string
		reportSizeBytes int64
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT dataset_version_id, rule_id, status, evaluated_at, evaluated_by, summary, report_object_key, report_sha256, report_size_bytes
		 FROM quality_evaluations
		 WHERE evaluation_id = $1`,
		evaluationID,
	).Scan(&versionID, &ruleID, &status, &evaluatedAt, &evaluatedBy, &summaryRaw, &reportObjectKey, &reportSHA256, &reportSizeBytes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, evaluation{
		EvaluationID:     evaluationID,
		DatasetVersionID: versionID,
		RuleID:           ruleID,
		Status:           status,
		EvaluatedAt:      evaluatedAt,
		EvaluatedBy:      evaluatedBy,
		Summary:          normalizeJSON(summaryRaw),
		ReportObjectKey:  reportObjectKey,
		ReportSHA256:     reportSHA256,
		ReportSizeBytes:  reportSizeBytes,
	})
}

type gateStatus struct {
	DatasetVersionID string    `json:"dataset_version_id"`
	RuleID           string    `json:"rule_id,omitempty"`
	Status           string    `json:"status"`
	EvaluationID     string    `json:"evaluation_id,omitempty"`
	EvaluatedAt      time.Time `json:"evaluated_at,omitempty"`
}

func (api *qualityAPI) handleGateStatus(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.PathValue("version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_version_id_required")
		return
	}

	var ruleID sql.NullString
	if err := api.db.QueryRowContext(
		r.Context(),
		`SELECT quality_rule_id FROM dataset_versions WHERE version_id = $1`,
		versionID,
	).Scan(&ruleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if strings.TrimSpace(ruleID.String) == "" {
		api.writeJSON(w, http.StatusOK, gateStatus{
			DatasetVersionID: versionID,
			Status:           "no_rule",
		})
		return
	}

	var (
		evaluationID string
		status       string
		evaluatedAt  time.Time
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT evaluation_id, status, evaluated_at
		 FROM quality_evaluations
		 WHERE dataset_version_id = $1 AND rule_id = $2
		 ORDER BY evaluated_at DESC
		 LIMIT 1`,
		versionID,
		strings.TrimSpace(ruleID.String),
	).Scan(&evaluationID, &status, &evaluatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeJSON(w, http.StatusOK, gateStatus{
				DatasetVersionID: versionID,
				RuleID:           strings.TrimSpace(ruleID.String),
				Status:           "not_evaluated",
			})
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, gateStatus{
		DatasetVersionID: versionID,
		RuleID:           strings.TrimSpace(ruleID.String),
		Status:           status,
		EvaluationID:     evaluationID,
		EvaluatedAt:      evaluatedAt,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 4<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}

func decodeJSONBytes(raw []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}

func (api *qualityAPI) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(body)
}

func (api *qualityAPI) writeError(w http.ResponseWriter, r *http.Request, status int, code string) {
	api.writeJSON(w, status, map[string]any{
		"error":      code,
		"request_id": r.Header.Get("X-Request-Id"),
	})
}

func normalizeJSON(raw []byte) json.RawMessage {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("{}")
	}
	return raw
}

func jsonFieldString(raw json.RawMessage, key string) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	v, ok := obj[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func requestIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}

func clampInt(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func nullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
