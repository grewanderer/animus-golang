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

// StepState represents a terminal step outcome.
type StepState string

const (
	StepStateSucceeded StepState = "succeeded"
	StepStateFailed    StepState = "failed"
	StepStateSkipped   StepState = "skipped"
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

// CanTransition enforces forward-only state progression.
func CanTransition(current, next RunState) bool {
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
