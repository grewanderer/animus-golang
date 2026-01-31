package state

import (
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

func TestDeriveRunState(t *testing.T) {
	plan := &domain.ExecutionPlan{
		RunID:     "run-1",
		ProjectID: "proj-1",
		Steps: []domain.ExecutionPlanStep{
			{Name: "a"},
			{Name: "b"},
		},
		Edges: []domain.ExecutionPlanEdge{
			{From: "a", To: "b"},
		},
	}

	tests := []struct {
		name       string
		plan       *domain.ExecutionPlan
		executions []repo.StepExecutionRecord
		want       domain.RunState
	}{
		{
			name: "no plan",
			plan: nil,
			want: domain.RunStateCreated,
		},
		{
			name: "plan no executions",
			plan: plan,
			want: domain.RunStatePlanned,
		},
		{
			name: "all succeeded",
			plan: plan,
			executions: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Succeeded"),
				stepRecord("b", 1, "Succeeded"),
			},
			want: domain.RunStateDryRunSucceeded,
		},
		{
			name: "failed step",
			plan: plan,
			executions: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Failed"),
			},
			want: domain.RunStateDryRunFailed,
		},
		{
			name: "skipped without failed ancestor",
			plan: plan,
			executions: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Succeeded"),
				stepRecord("b", 1, "Skipped"),
			},
			want: domain.RunStateDryRunFailed,
		},
		{
			name: "skipped with failed ancestor",
			plan: plan,
			executions: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Failed"),
				stepRecord("b", 1, "Skipped"),
			},
			want: domain.RunStateDryRunFailed,
		},
		{
			name: "partial execution",
			plan: plan,
			executions: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Retried"),
			},
			want: domain.RunStateDryRunRunning,
		},
	}

	for _, tc := range tests {
		if got := DeriveRunState(tc.plan, tc.executions); got != tc.want {
			t.Fatalf("%s: expected %s got %s", tc.name, tc.want, got)
		}
	}
}

func TestDeriveStepOutcome(t *testing.T) {
	tests := []struct {
		name         string
		records      []repo.StepExecutionRecord
		wantAttempts int
		wantStatus   domain.StepState
	}{
		{
			name:         "no attempts",
			wantAttempts: 0,
			wantStatus:   "",
		},
		{
			name: "succeeded",
			records: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Succeeded"),
			},
			wantAttempts: 1,
			wantStatus:   domain.StepStateSucceeded,
		},
		{
			name: "failed after retry",
			records: []repo.StepExecutionRecord{
				stepRecord("a", 1, "Retried"),
				stepRecord("a", 2, "Failed"),
			},
			wantAttempts: 2,
			wantStatus:   domain.StepStateFailed,
		},
	}

	for _, tc := range tests {
		attempts, status := DeriveStepOutcome(tc.records)
		if attempts != tc.wantAttempts || status != tc.wantStatus {
			t.Fatalf("%s: expected %d/%s got %d/%s", tc.name, tc.wantAttempts, tc.wantStatus, attempts, status)
		}
	}
}

func stepRecord(step string, attempt int, status string) repo.StepExecutionRecord {
	return repo.StepExecutionRecord{
		ProjectID: "proj-1",
		RunID:     "run-1",
		StepName:  step,
		Attempt:   attempt,
		Status:    status,
	}
}
