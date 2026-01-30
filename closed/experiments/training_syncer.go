package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
)

type trainingSyncer struct {
	logger   *slog.Logger
	db       *sql.DB
	executor trainingExecutor
	interval time.Duration
	batch    int
}

func startTrainingSyncer(ctx context.Context, logger *slog.Logger, db *sql.DB, executor trainingExecutor, interval time.Duration) {
	if db == nil || executor == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	s := &trainingSyncer{
		logger:   logger,
		db:       db,
		executor: executor,
		interval: interval,
		batch:    50,
	}

	go s.run(ctx)
}

func (s *trainingSyncer) run(ctx context.Context) {
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

func (s *trainingSyncer) syncOnce(ctx context.Context) {
	kind := strings.TrimSpace(s.executor.Kind())
	if kind == "" {
		return
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT run_id, executor, COALESCE(k8s_namespace,''), COALESCE(k8s_job_name,''), COALESCE(docker_container_id,'')
		 FROM experiment_run_executions e
		 WHERE executor = $1
		   AND NOT EXISTS (
			 SELECT 1
			 FROM experiment_run_state_events s
			 WHERE s.run_id = e.run_id
			   AND s.status IN ('succeeded','failed','canceled')
		   )
		 ORDER BY created_at ASC
		 LIMIT $2`,
		kind,
		s.batch,
	)
	if err != nil {
		s.log("sync query failed", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var exec trainingExecution
		if err := rows.Scan(&exec.RunID, &exec.Executor, &exec.K8sNamespace, &exec.K8sJobName, &exec.DockerContainer); err != nil {
			s.log("scan execution failed", "error", err)
			return
		}
		s.syncExecution(ctx, exec)
	}
	if err := rows.Err(); err != nil {
		s.log("sync rows error", "error", err)
	}
}

func (s *trainingSyncer) syncExecution(ctx context.Context, exec trainingExecution) {
	obs, err := s.executor.Inspect(ctx, exec)
	if err != nil {
		s.log("inspect failed", "run_id", exec.RunID, "error", err)
		return
	}

	status := strings.ToLower(strings.TrimSpace(obs.Status))
	switch status {
	case "pending", "":
		return
	case "running", "succeeded", "failed":
	default:
		s.log("unexpected training status", "run_id", exec.RunID, "status", obs.Status)
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
	details["executor"] = exec.Executor

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log("begin tx failed", "run_id", exec.RunID, "error", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := (&experimentsAPI{db: s.db}).insertRunStateEvent(ctx, tx, exec.RunID, status, now, details)
	if err != nil {
		s.log("insert state failed", "run_id", exec.RunID, "status", status, "error", err)
		return
	}
	if !inserted {
		_ = tx.Rollback()
		return
	}

	level := "info"
	eventMessage := "run " + status
	if status == "failed" {
		level = "error"
	}
	_ = (&experimentsAPI{db: s.db}).insertRunEvent(ctx, tx, exec.RunID, "system", level, eventMessage, details)

	action := "experiment_run." + status
	if status == "running" {
		action = "experiment_run.running"
	}
	_, err = auditlog.Insert(ctx, tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        "system",
		Action:       action,
		ResourceType: "experiment_run",
		ResourceID:   exec.RunID,
		Payload: map[string]any{
			"service": "experiments",
			"run_id":  exec.RunID,
			"status":  status,
			"details": details,
		},
	})
	if err != nil {
		s.log("insert audit failed", "run_id", exec.RunID, "status", status, "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.log("commit failed", "run_id", exec.RunID, "status", status, "error", err)
	}
}

func (s *trainingSyncer) log(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	fields := []any{"component", "training_syncer"}
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
