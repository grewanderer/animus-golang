package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	executor "github.com/animus-labs/animus-go/closed/internal/execution/executor"
	"github.com/animus-labs/animus-go/closed/internal/execution/executor/dryrun"
	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	"github.com/animus-labs/animus-go/closed/internal/repo/postgres"
)

type dryRunResponse struct {
	RunID    string                      `json:"runId"`
	Status   string                      `json:"status"`
	Existing bool                        `json:"existing"`
	Steps    []executor.DryRunStepResult `json:"steps"`
}

type dryRunExecutionsResponse struct {
	RunID      string                  `json:"runId"`
	Executions []stepExecutionResponse `json:"executions"`
}

type stepExecutionResponse struct {
	StepName     string          `json:"stepName"`
	Attempt      int             `json:"attempt"`
	Status       string          `json:"status"`
	StartedAt    time.Time       `json:"startedAt"`
	FinishedAt   *time.Time      `json:"finishedAt,omitempty"`
	ErrorCode    string          `json:"errorCode,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}

func (api *experimentsAPI) handleDryRun(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	projectID := strings.TrimSpace(r.PathValue("project_id"))
	runID := strings.TrimSpace(r.PathValue("run_id"))
	if projectID == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}
	if runID == "" {
		api.writeError(w, r, http.StatusBadRequest, "run_id_required")
		return
	}

	tx, err := api.db.BeginTx(r.Context(), nil)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	runStore := postgres.NewRunSpecStore(tx)
	planStore := postgres.NewPlanStore(tx)
	stepStore := postgres.NewStepExecutionStore(tx)
	if runStore == nil || planStore == nil || stepStore == nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	runRecord, err := runStore.GetRun(r.Context(), projectID, runID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	planRecord, err := planStore.GetPlan(r.Context(), projectID, runID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "plan_not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	execPlan, err := decodeExecutionPlan(planRecord.Plan)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "invalid_plan")
		return
	}

	exec := dryrun.New(stepStore)
	result, err := exec.DryRun(r.Context(), executor.DryRunInput{
		ProjectID: projectID,
		RunID:     runID,
		SpecHash:  runRecord.SpecHash,
		Plan:      execPlan,
	})
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "dry_run_failed")
		return
	}

	if !result.Existing {
		if err := api.appendDryRunAuditEvents(r, tx, identity.Subject, projectID, runID, runRecord.SpecHash, result); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, dryRunResponse{
		RunID:    runID,
		Status:   result.Status,
		Existing: result.Existing,
		Steps:    result.Steps,
	})
}

func (api *experimentsAPI) handleGetDryRun(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.PathValue("project_id"))
	runID := strings.TrimSpace(r.PathValue("run_id"))
	if projectID == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}
	if runID == "" {
		api.writeError(w, r, http.StatusBadRequest, "run_id_required")
		return
	}

	stepStore := postgres.NewStepExecutionStore(api.db)
	if stepStore == nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	records, err := stepStore.ListByRun(r.Context(), projectID, runID)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if len(records) == 0 {
		api.writeError(w, r, http.StatusNotFound, "not_found")
		return
	}

	executions := make([]stepExecutionResponse, 0, len(records))
	for _, record := range records {
		executions = append(executions, stepExecutionResponse{
			StepName:     record.StepName,
			Attempt:      record.Attempt,
			Status:       record.Status,
			StartedAt:    record.StartedAt,
			FinishedAt:   record.FinishedAt,
			ErrorCode:    strings.TrimSpace(record.ErrorCode),
			ErrorMessage: strings.TrimSpace(record.ErrorMessage),
			Result:       json.RawMessage(record.Result),
		})
	}

	api.writeJSON(w, http.StatusOK, dryRunExecutionsResponse{
		RunID:      runID,
		Executions: executions,
	})
}

func (api *experimentsAPI) appendDryRunAuditEvents(r *http.Request, q auditlog.QueryRower, actor, projectID, runID, specHash string, result executor.DryRunResult) error {
	now := time.Now().UTC()
	_, err := auditlog.Insert(r.Context(), q, auditlog.Event{
		OccurredAt:   now,
		Actor:        actor,
		Action:       "dry_run.started",
		ResourceType: "run",
		ResourceID:   runID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":    "experiments",
			"project_id": projectID,
			"run_id":     runID,
			"spec_hash":  specHash,
		},
	})
	if err != nil {
		return err
	}

	for _, attempt := range result.Attempts {
		_, err = auditlog.Insert(r.Context(), q, auditlog.Event{
			OccurredAt:   now,
			Actor:        actor,
			Action:       "dry_run.step.started",
			ResourceType: "step_execution",
			ResourceID:   runID,
			RequestID:    r.Header.Get("X-Request-Id"),
			IP:           requestIP(r.RemoteAddr),
			UserAgent:    r.UserAgent(),
			Payload: map[string]any{
				"service":    "experiments",
				"project_id": projectID,
				"run_id":     runID,
				"step_name":  attempt.StepName,
				"attempt":    attempt.Attempt,
				"status":     attempt.Status,
				"spec_hash":  specHash,
			},
		})
		if err != nil {
			return err
		}

		action := "dry_run.step.completed"
		switch attempt.Status {
		case dryrun.StatusRetried:
			action = "dry_run.step.retried"
		case dryrun.StatusSkipped:
			action = "dry_run.step.skipped"
		}
		_, err = auditlog.Insert(r.Context(), q, auditlog.Event{
			OccurredAt:   now,
			Actor:        actor,
			Action:       action,
			ResourceType: "step_execution",
			ResourceID:   runID,
			RequestID:    r.Header.Get("X-Request-Id"),
			IP:           requestIP(r.RemoteAddr),
			UserAgent:    r.UserAgent(),
			Payload: map[string]any{
				"service":    "experiments",
				"project_id": projectID,
				"run_id":     runID,
				"step_name":  attempt.StepName,
				"attempt":    attempt.Attempt,
				"status":     attempt.Status,
				"spec_hash":  specHash,
			},
		})
		if err != nil {
			return err
		}
	}

	finalAction := "dry_run.completed"
	if result.Status != dryrun.StatusSucceeded {
		finalAction = "dry_run.failed"
	}
	_, err = auditlog.Insert(r.Context(), q, auditlog.Event{
		OccurredAt:   now,
		Actor:        actor,
		Action:       finalAction,
		ResourceType: "run",
		ResourceID:   runID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":    "experiments",
			"project_id": projectID,
			"run_id":     runID,
			"spec_hash":  specHash,
			"status":     result.Status,
		},
	})
	return err
}

func decodeExecutionPlan(raw []byte) (domain.ExecutionPlan, error) {
	var payload executionPlanPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.ExecutionPlan{}, err
	}
	steps := make([]domain.ExecutionPlanStep, 0, len(payload.Steps))
	for _, step := range payload.Steps {
		steps = append(steps, domain.ExecutionPlanStep{
			Name: step.Name,
			RetryPolicy: domain.PipelineRetryPolicy{
				MaxAttempts: step.RetryPolicy.MaxAttempts,
				Backoff: domain.PipelineBackoff{
					Type:           step.RetryPolicy.Backoff.Type,
					InitialSeconds: step.RetryPolicy.Backoff.InitialSeconds,
					MaxSeconds:     step.RetryPolicy.Backoff.MaxSeconds,
					Multiplier:     step.RetryPolicy.Backoff.Multiplier,
				},
			},
			AttemptStart: step.AttemptStart,
		})
	}
	edges := make([]domain.ExecutionPlanEdge, 0, len(payload.Edges))
	for _, edge := range payload.Edges {
		edges = append(edges, domain.ExecutionPlanEdge{
			From: edge.From,
			To:   edge.To,
		})
	}
	return domain.ExecutionPlan{
		RunID:     payload.RunID,
		ProjectID: payload.ProjectID,
		Steps:     steps,
		Edges:     edges,
	}, nil
}
