package dryrun

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	executor "github.com/animus-labs/animus-go/closed/internal/execution/executor"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

func TestDryRunDeterministic(t *testing.T) {
	plan := samplePlan(2, 1)
	input := executor.DryRunInput{
		ProjectID: "proj-1",
		RunID:     "run-1",
		SpecHash:  "spec-hash",
		Plan:      plan,
	}

	firstRepo := newMemoryRepo()
	secondRepo := newMemoryRepo()
	fixed := func() time.Time { return time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC) }

	first := New(firstRepo)
	first.now = fixed
	second := New(secondRepo)
	second.now = fixed

	firstResult, err := first.DryRun(context.Background(), input)
	if err != nil {
		t.Fatalf("first dry run: %v", err)
	}
	secondResult, err := second.DryRun(context.Background(), input)
	if err != nil {
		t.Fatalf("second dry run: %v", err)
	}

	if firstResult.Status != secondResult.Status {
		t.Fatalf("expected deterministic status, got %s vs %s", firstResult.Status, secondResult.Status)
	}
	if len(firstResult.Steps) != len(secondResult.Steps) {
		t.Fatalf("expected same step count")
	}
	for i := range firstResult.Steps {
		if firstResult.Steps[i] != secondResult.Steps[i] {
			t.Fatalf("expected deterministic step results, got %+v vs %+v", firstResult.Steps[i], secondResult.Steps[i])
		}
	}
}

func TestDryRunRetryBehavior(t *testing.T) {
	plan := samplePlan(1, 3)
	repo := newMemoryRepo()
	exec := New(repo)
	exec.now = func() time.Time { return time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC) }
	exec.decide = func(specHash, runID, stepName string, attempt int) float64 {
		if attempt < 3 {
			return 0.99
		}
		return 0.01
	}

	result, err := exec.DryRun(context.Background(), executor.DryRunInput{
		ProjectID: "proj-1",
		RunID:     "run-1",
		SpecHash:  "spec-hash",
		Plan:      plan,
	})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if result.Status != StatusSucceeded {
		t.Fatalf("expected success, got %s", result.Status)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("expected one step result")
	}
	if result.Steps[0].Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", result.Steps[0].Attempts)
	}
}

func TestDryRunIdempotent(t *testing.T) {
	plan := samplePlan(1, 1)
	repo := newMemoryRepo()
	exec := New(repo)

	input := executor.DryRunInput{
		ProjectID: "proj-1",
		RunID:     "run-1",
		SpecHash:  "spec-hash",
		Plan:      plan,
	}
	first, err := exec.DryRun(context.Background(), input)
	if err != nil {
		t.Fatalf("first dry run: %v", err)
	}
	if first.Existing {
		t.Fatalf("expected first run to be new")
	}
	before := repo.count()

	second, err := exec.DryRun(context.Background(), input)
	if err != nil {
		t.Fatalf("second dry run: %v", err)
	}
	if !second.Existing {
		t.Fatalf("expected second run to be existing")
	}
	after := repo.count()
	if before != after {
		t.Fatalf("expected no new inserts, got %d before %d after", before, after)
	}
}

func samplePlan(stepCount int, maxAttempts int) domain.ExecutionPlan {
	steps := make([]domain.ExecutionPlanStep, 0, stepCount)
	for i := 0; i < stepCount; i++ {
		name := string(rune('a' + i))
		steps = append(steps, domain.ExecutionPlanStep{
			Name: name,
			RetryPolicy: domain.PipelineRetryPolicy{
				MaxAttempts: maxAttempts,
				Backoff: domain.PipelineBackoff{
					Type:           "fixed",
					InitialSeconds: 0,
					MaxSeconds:     0,
					Multiplier:     1,
				},
			},
			AttemptStart: 1,
		})
	}
	return domain.ExecutionPlan{
		RunID:     "run-1",
		ProjectID: "proj-1",
		Steps:     steps,
	}
}

type memoryRepo struct {
	records map[string]repo.StepExecutionRecord
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{records: map[string]repo.StepExecutionRecord{}}
}

func (m *memoryRepo) InsertAttempt(ctx context.Context, record repo.StepExecutionRecord) (repo.StepExecutionRecord, bool, error) {
	key := record.ProjectID + "/" + record.RunID + "/" + record.StepName + "/" + itoa(record.Attempt)
	if existing, ok := m.records[key]; ok {
		return existing, false, nil
	}
	m.records[key] = record
	return record, true, nil
}

func (m *memoryRepo) ListByRun(ctx context.Context, projectID, runID string) ([]repo.StepExecutionRecord, error) {
	out := make([]repo.StepExecutionRecord, 0)
	for _, record := range m.records {
		if record.ProjectID == projectID && record.RunID == runID {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].StartedAt.Equal(out[j].StartedAt) {
			return out[i].StartedAt.Before(out[j].StartedAt)
		}
		if out[i].StepName != out[j].StepName {
			return out[i].StepName < out[j].StepName
		}
		return out[i].Attempt < out[j].Attempt
	})
	return out, nil
}

func (m *memoryRepo) count() int {
	return len(m.records)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	buf := make([]byte, 0, 12)
	for value > 0 {
		buf = append(buf, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return sign + string(buf)
}
