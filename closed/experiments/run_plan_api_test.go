package main

import (
	"encoding/json"
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

func TestMarshalExecutionPlan(t *testing.T) {
	plan := domain.ExecutionPlan{
		RunID:     "run-1",
		ProjectID: "proj-1",
		Steps: []domain.ExecutionPlanStep{
			{
				Name: "step-b",
				RetryPolicy: domain.PipelineRetryPolicy{
					MaxAttempts: 2,
					Backoff: domain.PipelineBackoff{
						Type:           "fixed",
						InitialSeconds: 1,
						MaxSeconds:     2,
						Multiplier:     1,
					},
				},
				AttemptStart: 1,
			},
			{
				Name: "step-a",
				RetryPolicy: domain.PipelineRetryPolicy{
					MaxAttempts: 1,
					Backoff: domain.PipelineBackoff{
						Type:           "fixed",
						InitialSeconds: 0,
						MaxSeconds:     0,
						Multiplier:     1,
					},
				},
				AttemptStart: 1,
			},
		},
		Edges: []domain.ExecutionPlanEdge{
			{From: "step-b", To: "step-a"},
		},
	}

	raw, err := marshalExecutionPlan(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}

	var decoded executionPlanPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}

	if decoded.RunID != plan.RunID || decoded.ProjectID != plan.ProjectID {
		t.Fatalf("unexpected ids: %+v", decoded)
	}
	if len(decoded.Steps) != 2 || decoded.Steps[0].Name != "step-b" || decoded.Steps[1].Name != "step-a" {
		t.Fatalf("unexpected step order: %+v", decoded.Steps)
	}
	if decoded.Steps[0].AttemptStart != 1 {
		t.Fatalf("unexpected attempt start: %d", decoded.Steps[0].AttemptStart)
	}
	if len(decoded.Edges) != 1 || decoded.Edges[0].From != "step-b" || decoded.Edges[0].To != "step-a" {
		t.Fatalf("unexpected edges: %+v", decoded.Edges)
	}
}
