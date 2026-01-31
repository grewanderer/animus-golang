package state

import (
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

// DeriveRunState computes the deterministic run state from plan and step executions.
func DeriveRunState(plan *domain.ExecutionPlan, executions []repo.StepExecutionRecord) domain.RunState {
	if plan == nil {
		return domain.RunStateCreated
	}
	if len(plan.Steps) == 0 {
		return domain.RunStatePlanned
	}
	if len(executions) == 0 {
		return domain.RunStatePlanned
	}

	byStep := groupByStep(executions)
	failedSteps := map[string]struct{}{}
	skippedSteps := map[string]struct{}{}
	incomplete := false

	for _, step := range plan.Steps {
		stepName := strings.TrimSpace(step.Name)
		if stepName == "" {
			continue
		}
		attempts, status := DeriveStepOutcome(byStep[stepName])
		if attempts == 0 || status == "" {
			incomplete = true
			continue
		}
		switch status {
		case domain.StepStateFailed:
			failedSteps[stepName] = struct{}{}
		case domain.StepStateSkipped:
			skippedSteps[stepName] = struct{}{}
		}
	}

	if len(failedSteps) > 0 {
		return domain.RunStateDryRunFailed
	}
	if incomplete {
		return domain.RunStateDryRunRunning
	}

	if len(skippedSteps) > 0 {
		if !skipsHaveFailedAncestor(plan, failedSteps, skippedSteps) {
			return domain.RunStateDryRunFailed
		}
	}
	return domain.RunStateDryRunSucceeded
}

// DeriveStepOutcome returns attempts count and terminal status (empty if not terminal).
func DeriveStepOutcome(executions []repo.StepExecutionRecord) (int, domain.StepState) {
	if len(executions) == 0 {
		return 0, ""
	}
	maxAttempt := 0
	finalStatus := ""
	for _, record := range executions {
		if record.Attempt > maxAttempt {
			maxAttempt = record.Attempt
			finalStatus = record.Status
		}
	}
	if maxAttempt == 0 {
		return 0, ""
	}
	if state, ok := mapStepStatus(finalStatus); ok {
		return maxAttempt, state
	}
	return maxAttempt, ""
}

func mapStepStatus(status string) (domain.StepState, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "succeeded":
		return domain.StepStateSucceeded, true
	case "failed":
		return domain.StepStateFailed, true
	case "skipped":
		return domain.StepStateSkipped, true
	default:
		return "", false
	}
}

func groupByStep(executions []repo.StepExecutionRecord) map[string][]repo.StepExecutionRecord {
	out := make(map[string][]repo.StepExecutionRecord)
	for _, record := range executions {
		step := strings.TrimSpace(record.StepName)
		if step == "" {
			continue
		}
		out[step] = append(out[step], record)
	}
	return out
}

func skipsHaveFailedAncestor(plan *domain.ExecutionPlan, failedSteps, skippedSteps map[string]struct{}) bool {
	if len(skippedSteps) == 0 {
		return true
	}
	deps := reverseDependencies(plan.Edges)
	for step := range skippedSteps {
		if !hasFailedAncestor(step, deps, failedSteps, map[string]struct{}{}) {
			return false
		}
	}
	return true
}

func reverseDependencies(edges []domain.ExecutionPlanEdge) map[string][]string {
	out := make(map[string][]string)
	for _, edge := range edges {
		from := strings.TrimSpace(edge.From)
		to := strings.TrimSpace(edge.To)
		if from == "" || to == "" {
			continue
		}
		out[to] = append(out[to], from)
	}
	return out
}

func hasFailedAncestor(step string, deps map[string][]string, failedSteps map[string]struct{}, visited map[string]struct{}) bool {
	if _, ok := visited[step]; ok {
		return false
	}
	visited[step] = struct{}{}
	for _, parent := range deps[step] {
		if _, ok := failedSteps[parent]; ok {
			return true
		}
		if hasFailedAncestor(parent, deps, failedSteps, visited) {
			return true
		}
	}
	return false
}
