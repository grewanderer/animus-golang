package main

import (
	"reflect"
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

func TestDecodeExecutionPlan(t *testing.T) {
	plan := domain.ExecutionPlan{
		RunID:     "run-1",
		ProjectID: "proj-1",
		Steps: []domain.ExecutionPlanStep{
			{
				Name: "step-a",
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
		},
		Edges: []domain.ExecutionPlanEdge{
			{From: "step-a", To: "step-b"},
		},
	}

	raw, err := marshalExecutionPlan(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	decoded, err := decodeExecutionPlan(raw)
	if err != nil {
		t.Fatalf("decode plan: %v", err)
	}

	if decoded.RunID != plan.RunID || decoded.ProjectID != plan.ProjectID {
		t.Fatalf("unexpected ids: %+v", decoded)
	}
	if !reflect.DeepEqual(decoded.Steps, plan.Steps) {
		t.Fatalf("steps mismatch: %+v vs %+v", decoded.Steps, plan.Steps)
	}
	if !reflect.DeepEqual(decoded.Edges, plan.Edges) {
		t.Fatalf("edges mismatch: %+v vs %+v", decoded.Edges, plan.Edges)
	}
}
