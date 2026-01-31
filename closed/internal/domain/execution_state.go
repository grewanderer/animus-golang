package domain

import "strings"

// RunState represents the derived execution state of a run.
type RunState string

const (
	RunStateCreated         RunState = "created"
	RunStatePlanned         RunState = "planned"
	RunStateDryRunRunning   RunState = "dryrun_running"
	RunStateDryRunSucceeded RunState = "dryrun_succeeded"
	RunStateDryRunFailed    RunState = "dryrun_failed"
)

// StepOutcome represents a terminal step outcome.
type StepOutcome string

const (
	StepOutcomeSucceeded StepOutcome = "succeeded"
	StepOutcomeFailed    StepOutcome = "failed"
	StepOutcomeSkipped   StepOutcome = "skipped"
)

// StepState is kept for backward compatibility with earlier call sites.
type StepState = StepOutcome

const (
	StepStateSucceeded StepState = StepOutcomeSucceeded
	StepStateFailed    StepState = StepOutcomeFailed
	StepStateSkipped   StepState = StepOutcomeSkipped
)

// NormalizeRunState maps free-form status values to canonical run states.
func NormalizeRunState(value string) RunState {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(RunStateCreated), "pending":
		return RunStateCreated
	case string(RunStatePlanned):
		return RunStatePlanned
	case string(RunStateDryRunRunning):
		return RunStateDryRunRunning
	case string(RunStateDryRunSucceeded):
		return RunStateDryRunSucceeded
	case string(RunStateDryRunFailed):
		return RunStateDryRunFailed
	default:
		return ""
	}
}

// CanTransitionRunState enforces forward-only state progression.
func CanTransitionRunState(current, next RunState) bool {
	if current == "" || next == "" {
		return false
	}
	if current == next {
		return true
	}
	return runStateOrder(current) < runStateOrder(next)
}

func runStateOrder(state RunState) int {
	switch state {
	case RunStateCreated:
		return 1
	case RunStatePlanned:
		return 2
	case RunStateDryRunRunning:
		return 3
	case RunStateDryRunSucceeded, RunStateDryRunFailed:
		return 4
	default:
		return 0
	}
}
