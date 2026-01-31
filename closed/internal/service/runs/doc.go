// Package runs implements the execution state machine for project-scoped runs.
//
// States:
//   - created -> planned -> dryrun_running -> dryrun_succeeded | dryrun_failed
//
// Transitions are derived from the stored ExecutionPlan and step_executions via
// DeriveAndPersistWithAudit. Read-only callers should use Derive, which does not
// mutate persisted status. Explicit transitions (e.g. dryrun_running) must be
// applied through the service to enforce invariants.
//
// Auditing:
//   - Successful transitions emit exactly one run-level audit event.
//   - Rejected transitions do not emit audit events (callers should handle errors).
package runs
