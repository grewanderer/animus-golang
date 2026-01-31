package runs

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/execution/plan"
	"github.com/animus-labs/animus-go/closed/internal/execution/state"
	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

type Service struct {
	runs  repo.RunRepository
	plans repo.PlanRepository
	steps repo.StepExecutionRepository
}

type AuditInfo struct {
	Actor     string
	RequestID string
	UserAgent string
	IP        net.IP
	Service   string
}

func New(runRepo repo.RunRepository, planRepo repo.PlanRepository, stepRepo repo.StepExecutionRepository) *Service {
	if runRepo == nil || planRepo == nil || stepRepo == nil {
		return nil
	}
	return &Service{
		runs:  runRepo,
		plans: planRepo,
		steps: stepRepo,
	}
}

// DeriveAndPersist computes the derived run state and persists it if needed.
func (s *Service) DeriveAndPersist(ctx context.Context, projectID, runID string) (repo.RunRecord, domain.RunState, domain.RunState, error) {
	runRecord, err := s.runs.GetRun(ctx, projectID, runID)
	if err != nil {
		return repo.RunRecord{}, "", "", err
	}
	prev := domain.NormalizeRunState(runRecord.Status)
	if prev == "" {
		prev = domain.RunStateCreated
	}
	planSpec, err := s.loadPlan(ctx, projectID, runID)
	if err != nil {
		return repo.RunRecord{}, "", "", err
	}
	stepExecutions, err := s.steps.ListByRun(ctx, projectID, runID)
	if err != nil {
		return repo.RunRecord{}, "", "", err
	}
	derived := state.DeriveRunState(planSpec, stepExecutions)
	if err := s.runs.UpdateDerivedStatus(ctx, projectID, runID, derived); err != nil {
		return repo.RunRecord{}, "", "", err
	}
	return runRecord, prev, derived, nil
}

// MarkDryRunRunning transitions a run to dryrun_running after verifying a plan exists.
func (s *Service) MarkDryRunRunning(ctx context.Context, projectID, runID string) (domain.RunState, error) {
	runRecord, err := s.runs.GetRun(ctx, projectID, runID)
	if err != nil {
		return "", err
	}
	prev := domain.NormalizeRunState(runRecord.Status)
	if prev == "" {
		prev = domain.RunStateCreated
	}
	planSpec, err := s.loadPlan(ctx, projectID, runID)
	if err != nil {
		return "", err
	}
	if planSpec == nil {
		return "", errors.New("plan required for dry-run")
	}
	if err := s.runs.UpdateDerivedStatus(ctx, projectID, runID, domain.RunStateDryRunRunning); err != nil {
		return "", err
	}
	return prev, nil
}

// AppendRunTransitionAudit emits audit events aligned with run state transitions.
func (s *Service) AppendRunTransitionAudit(ctx context.Context, q auditlog.QueryRower, info AuditInfo, projectID, runID, specHash string, from, to domain.RunState) error {
	if q == nil {
		return errors.New("audit queryer is required")
	}
	if strings.TrimSpace(info.Actor) == "" {
		return errors.New("audit actor is required")
	}
	if from == to {
		return nil
	}

	var action string
	switch to {
	case domain.RunStateDryRunRunning:
		action = "dry_run.started"
	case domain.RunStateDryRunSucceeded:
		action = "dry_run.completed"
	case domain.RunStateDryRunFailed:
		action = "dry_run.failed"
	default:
		return nil
	}

	_, err := auditlog.Insert(ctx, q, auditlog.Event{
		Actor:        info.Actor,
		Action:       action,
		ResourceType: "run",
		ResourceID:   runID,
		RequestID:    info.RequestID,
		IP:           info.IP,
		UserAgent:    info.UserAgent,
		Payload: map[string]any{
			"service":    strings.TrimSpace(info.Service),
			"project_id": projectID,
			"run_id":     runID,
			"spec_hash":  specHash,
			"from":       string(from),
			"to":         string(to),
		},
	})
	return err
}

func (s *Service) loadPlan(ctx context.Context, projectID, runID string) (*domain.ExecutionPlan, error) {
	planRecord, err := s.plans.GetPlan(ctx, projectID, runID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	parsed, err := plan.UnmarshalExecutionPlan(planRecord.Plan)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
