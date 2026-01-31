package runs

import (
	"context"
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/execution/plan"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

func TestDeriveAndPersist(t *testing.T) {
	runID := "run-1"
	projectID := "proj-1"
	runRepo := newFakeRunRepo(runID, projectID, string(domain.RunStateCreated))
	planRepo := &fakePlanRepo{plans: map[string]repo.PlanRecord{}}
	stepRepo := &fakeStepRepo{}

	execPlan := domain.ExecutionPlan{
		RunID:     runID,
		ProjectID: projectID,
		Steps: []domain.ExecutionPlanStep{
			{Name: "a"},
			{Name: "b"},
		},
	}
	planJSON, err := plan.MarshalExecutionPlan(execPlan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	planRepo.plans[runID] = repo.PlanRecord{RunID: runID, ProjectID: projectID, Plan: planJSON}

	service := New(runRepo, planRepo, stepRepo)
	if service == nil {
		t.Fatalf("expected service")
	}

	_, prev, derived, err := service.DeriveAndPersist(context.Background(), projectID, runID)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}
	if prev != domain.RunStateCreated || derived != domain.RunStatePlanned {
		t.Fatalf("unexpected state transition: %s -> %s", prev, derived)
	}
	if runRepo.records[runID].Status != string(domain.RunStatePlanned) {
		t.Fatalf("expected persisted planned status")
	}

	stepRepo.executions = []repo.StepExecutionRecord{
		{ProjectID: projectID, RunID: runID, StepName: "a", Attempt: 1, Status: "Succeeded"},
		{ProjectID: projectID, RunID: runID, StepName: "b", Attempt: 1, Status: "Succeeded"},
	}
	_, prev, derived, err = service.DeriveAndPersist(context.Background(), projectID, runID)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}
	if prev != domain.RunStatePlanned || derived != domain.RunStateDryRunSucceeded {
		t.Fatalf("unexpected state transition: %s -> %s", prev, derived)
	}
	if runRepo.records[runID].Status != string(domain.RunStateDryRunSucceeded) {
		t.Fatalf("expected persisted dryrun_succeeded status")
	}
}

func TestMarkDryRunRunningRequiresPlan(t *testing.T) {
	runRepo := newFakeRunRepo("run-1", "proj-1", string(domain.RunStateCreated))
	service := New(runRepo, &fakePlanRepo{plans: map[string]repo.PlanRecord{}}, &fakeStepRepo{})
	if _, err := service.MarkDryRunRunning(context.Background(), "proj-1", "run-1"); err == nil {
		t.Fatalf("expected error when plan missing")
	}
}

type fakeRunRepo struct {
	records map[string]repo.RunRecord
}

func newFakeRunRepo(runID, projectID, status string) *fakeRunRepo {
	return &fakeRunRepo{
		records: map[string]repo.RunRecord{
			runID: {
				ID:        runID,
				ProjectID: projectID,
				Status:    status,
			},
		},
	}
}

func (f *fakeRunRepo) CreateRun(ctx context.Context, projectID, idempotencyKey string, pipelineSpecJSON, runSpecJSON []byte, specHash string) (repo.RunRecord, bool, error) {
	return repo.RunRecord{}, false, nil
}

func (f *fakeRunRepo) GetRun(ctx context.Context, projectID, id string) (repo.RunRecord, error) {
	record, ok := f.records[id]
	if !ok || record.ProjectID != projectID {
		return repo.RunRecord{}, repo.ErrNotFound
	}
	return record, nil
}

func (f *fakeRunRepo) UpdateDerivedStatus(ctx context.Context, projectID, runID string, status domain.RunState) error {
	record, ok := f.records[runID]
	if !ok || record.ProjectID != projectID {
		return repo.ErrNotFound
	}
	current := domain.NormalizeRunState(record.Status)
	if current == "" {
		current = domain.RunStateCreated
	}
	if !domain.CanTransition(current, status) {
		return repo.ErrNotFound
	}
	record.Status = string(status)
	f.records[runID] = record
	return nil
}

type fakePlanRepo struct {
	plans map[string]repo.PlanRecord
}

func (f *fakePlanRepo) UpsertPlan(ctx context.Context, projectID, runID string, planJSON []byte) (repo.PlanRecord, error) {
	return repo.PlanRecord{}, nil
}

func (f *fakePlanRepo) GetPlan(ctx context.Context, projectID, runID string) (repo.PlanRecord, error) {
	record, ok := f.plans[runID]
	if !ok || record.ProjectID != projectID {
		return repo.PlanRecord{}, repo.ErrNotFound
	}
	return record, nil
}

type fakeStepRepo struct {
	executions []repo.StepExecutionRecord
}

func (f *fakeStepRepo) InsertAttempt(ctx context.Context, record repo.StepExecutionRecord) (repo.StepExecutionRecord, bool, error) {
	return repo.StepExecutionRecord{}, false, nil
}

func (f *fakeStepRepo) ListByRun(ctx context.Context, projectID, runID string) ([]repo.StepExecutionRecord, error) {
	out := make([]repo.StepExecutionRecord, 0)
	for _, record := range f.executions {
		if record.ProjectID == projectID && record.RunID == runID {
			out = append(out, record)
		}
	}
	return out, nil
}
