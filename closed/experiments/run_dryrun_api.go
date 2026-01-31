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
	"github.com/animus-labs/animus-go/closed/internal/execution/plan"
	"github.com/animus-labs/animus-go/closed/internal/execution/state"
	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	"github.com/animus-labs/animus-go/closed/internal/repo/postgres"
	"github.com/animus-labs/animus-go/closed/internal/service/runs"
)

type dryRunResponse struct {
	RunID    string                      `json:"runId"`
	Status   string                      `json:"status"`
	State    string                      `json:"state"`
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

	stateSvc := runs.New(runStore, planStore, stepStore)
	if stateSvc == nil {
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

	execPlan, err := plan.UnmarshalExecutionPlan(planRecord.Plan)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "invalid_plan")
		return
	}

	auditInfo := runs.AuditInfo{
		Actor:     identity.Subject,
		RequestID: r.Header.Get("X-Request-Id"),
		UserAgent: r.UserAgent(),
		IP:        requestIP(r.RemoteAddr),
		Service:   "experiments",
	}

	_, prevState, derivedState, err := stateSvc.DeriveAndPersist(r.Context(), projectID, runID)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if derivedState == domain.RunStateDryRunSucceeded || derivedState == domain.RunStateDryRunFailed {
		if err := stateSvc.AppendRunTransitionAudit(r.Context(), tx, auditInfo, projectID, runID, runRecord.SpecHash, prevState, derivedState); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
			return
		}
		records, err := stepStore.ListByRun(r.Context(), projectID, runID)
		if err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		if err := tx.Commit(); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		api.writeJSON(w, http.StatusOK, dryRunResponse{
			RunID:    runID,
			Status:   dryRunStatusFromState(derivedState),
			State:    string(derivedState),
			Existing: true,
			Steps:    buildDryRunSummary(execPlan, records),
		})
		return
	}

	if len(execPlan.Steps) == 0 {
		records, err := stepStore.ListByRun(r.Context(), projectID, runID)
		if err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		if err := tx.Commit(); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		api.writeJSON(w, http.StatusOK, dryRunResponse{
			RunID:    runID,
			Status:   dryRunStatusFromState(derivedState),
			State:    string(derivedState),
			Existing: true,
			Steps:    buildDryRunSummary(execPlan, records),
		})
		return
	}

	prevRunning, err := stateSvc.MarkDryRunRunning(r.Context(), projectID, runID)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if err := stateSvc.AppendRunTransitionAudit(r.Context(), tx, auditInfo, projectID, runID, runRecord.SpecHash, prevRunning, domain.RunStateDryRunRunning); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
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

	if len(result.Attempts) > 0 {
		if err := api.appendDryRunStepAuditEvents(r, tx, identity.Subject, projectID, runID, runRecord.SpecHash, result.Attempts); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
			return
		}
	}

	_, prevFinal, derivedFinal, err := stateSvc.DeriveAndPersist(r.Context(), projectID, runID)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if err := stateSvc.AppendRunTransitionAudit(r.Context(), tx, auditInfo, projectID, runID, runRecord.SpecHash, prevFinal, derivedFinal); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
		return
	}

	records, err := stepStore.ListByRun(r.Context(), projectID, runID)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	if err := tx.Commit(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, dryRunResponse{
		RunID:    runID,
		Status:   dryRunStatusFromState(derivedFinal),
		State:    string(derivedFinal),
		Existing: result.Existing,
		Steps:    buildDryRunSummary(execPlan, records),
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

func (api *experimentsAPI) appendDryRunStepAuditEvents(r *http.Request, q auditlog.QueryRower, actor, projectID, runID, specHash string, attempts []executor.DryRunAttempt) error {
	if len(attempts) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, attempt := range attempts {
		_, err := auditlog.Insert(r.Context(), q, auditlog.Event{
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
	return nil
}

func buildDryRunSummary(plan domain.ExecutionPlan, records []repo.StepExecutionRecord) []executor.DryRunStepResult {
	byStep := make(map[string][]repo.StepExecutionRecord)
	for _, record := range records {
		stepName := strings.TrimSpace(record.StepName)
		if stepName == "" {
			continue
		}
		byStep[stepName] = append(byStep[stepName], record)
	}

	results := make([]executor.DryRunStepResult, 0, len(plan.Steps))
	for _, step := range plan.Steps {
		stepName := strings.TrimSpace(step.Name)
		if stepName == "" {
			continue
		}
		attempts, outcome := state.DeriveStepOutcome(byStep[stepName])
		if attempts < 1 {
			attempts = 1
		}
		status := dryRunStatusForStep(outcome)
		if status == "" {
			status = dryrun.StatusFailed
		}
		results = append(results, executor.DryRunStepResult{
			Name:     stepName,
			Attempts: attempts,
			Status:   status,
		})
	}
	return results
}

func dryRunStatusForStep(outcome domain.StepState) string {
	switch outcome {
	case domain.StepStateSucceeded:
		return dryrun.StatusSucceeded
	case domain.StepStateFailed:
		return dryrun.StatusFailed
	case domain.StepStateSkipped:
		return dryrun.StatusSkipped
	default:
		return ""
	}
}

func dryRunStatusFromState(state domain.RunState) string {
	switch state {
	case domain.RunStateDryRunSucceeded:
		return dryrun.StatusSucceeded
	case domain.RunStateDryRunFailed:
		return dryrun.StatusFailed
	case domain.RunStateDryRunRunning:
		return "Running"
	case domain.RunStatePlanned:
		return "Planned"
	case domain.RunStateCreated:
		return "Created"
	default:
		return ""
	}
}
