package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/execution/plan"
	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	"github.com/animus-labs/animus-go/closed/internal/repo/postgres"
)

type planSummaryResponse struct {
	RunID     string   `json:"runId"`
	StepCount int      `json:"stepCount"`
	StepNames []string `json:"stepNames"`
}

func (api *experimentsAPI) handlePlanRun(w http.ResponseWriter, r *http.Request) {
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

	runStore := postgres.NewRunSpecStore(api.db)
	if runStore == nil {
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

	pipelineSpec, err := decodePipelineSpec(runRecord.PipelineSpec)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "invalid_pipeline_spec")
		return
	}

	execPlan, err := plan.BuildPlan(pipelineSpec, runID, projectID)
	if err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_pipeline_spec")
		return
	}

	planJSON, err := marshalExecutionPlan(execPlan)
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

	planStore := postgres.NewPlanStore(tx)
	if planStore == nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	planRecord, err := planStore.UpsertPlan(r.Context(), projectID, runID, planJSON)
	if err != nil {
		api.writeError(w, r, http.StatusConflict, "plan_conflict")
		return
	}

	_, err = auditlog.Insert(r.Context(), tx, auditlog.Event{
		OccurredAt:   time.Now().UTC(),
		Actor:        identity.Subject,
		Action:       "execution.planned",
		ResourceType: "execution_plan",
		ResourceID:   planRecord.ID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":    "experiments",
			"project_id": projectID,
			"run_id":     runID,
			"spec_hash":  runRecord.SpecHash,
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

	stepNames := make([]string, 0, len(execPlan.Steps))
	for _, step := range execPlan.Steps {
		stepNames = append(stepNames, step.Name)
	}
	api.writeJSON(w, http.StatusOK, planSummaryResponse{
		RunID:     runID,
		StepCount: len(execPlan.Steps),
		StepNames: stepNames,
	})
}

func (api *experimentsAPI) handleGetRunPlan(w http.ResponseWriter, r *http.Request) {
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

	planStore := postgres.NewPlanStore(api.db)
	if planStore == nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	record, err := planStore.GetPlan(r.Context(), projectID, runID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(record.Plan)
}

func marshalExecutionPlan(plan domain.ExecutionPlan) ([]byte, error) {
	payload := executionPlanPayload{
		RunID:     plan.RunID,
		ProjectID: plan.ProjectID,
		Steps:     make([]executionPlanStepPayload, 0, len(plan.Steps)),
		Edges:     make([]executionPlanEdgePayload, 0, len(plan.Edges)),
	}
	for _, step := range plan.Steps {
		payload.Steps = append(payload.Steps, executionPlanStepPayload{
			Name:         step.Name,
			RetryPolicy:  retryPolicyPayloadFromDomain(step.RetryPolicy),
			AttemptStart: step.AttemptStart,
		})
	}
	for _, edge := range plan.Edges {
		payload.Edges = append(payload.Edges, executionPlanEdgePayload{
			From: edge.From,
			To:   edge.To,
		})
	}
	return json.Marshal(payload)
}

type executionPlanPayload struct {
	RunID     string                     `json:"runId"`
	ProjectID string                     `json:"projectId"`
	Steps     []executionPlanStepPayload `json:"steps"`
	Edges     []executionPlanEdgePayload `json:"edges"`
}

type executionPlanStepPayload struct {
	Name         string             `json:"name"`
	RetryPolicy  retryPolicyPayload `json:"retryPolicy"`
	AttemptStart int                `json:"attemptStart"`
}

type executionPlanEdgePayload struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type retryPolicyPayload struct {
	MaxAttempts int            `json:"maxAttempts"`
	Backoff     backoffPayload `json:"backoff"`
}

type backoffPayload struct {
	Type           string  `json:"type"`
	InitialSeconds int     `json:"initialSeconds"`
	MaxSeconds     int     `json:"maxSeconds"`
	Multiplier     float64 `json:"multiplier"`
}

func retryPolicyPayloadFromDomain(policy domain.PipelineRetryPolicy) retryPolicyPayload {
	return retryPolicyPayload{
		MaxAttempts: policy.MaxAttempts,
		Backoff: backoffPayload{
			Type:           policy.Backoff.Type,
			InitialSeconds: policy.Backoff.InitialSeconds,
			MaxSeconds:     policy.Backoff.MaxSeconds,
			Multiplier:     policy.Backoff.Multiplier,
		},
	}
}
