# Execution model

Control, not a demo. Execution is contract-first and deterministic.

## Scope
- PipelineSpec is a reusable template.
- RunSpec is the immutable execution snapshot.
- ExecutionPlan is the deterministic DAG ordering.
- Dry-run simulates execution without running user code.
- Run state is derived from plan presence and step executions.

## Contracts

PipelineSpec
- Declares steps, dependencies, resources, and retry policy.
- References datasets by datasetRef (logical alias).
- Does not bind dataset versions.

RunSpec
- Binds datasetRef -> datasetVersionId.
- Binds codeRef (repoUrl + commitSha).
- Binds envLock (envHash, optional template id, image digests).
- Immutable once persisted.

ExecutionPlan
- Deterministic topological ordering of steps.
- Retry policy carried per step.
- Serializable and auditable.

## Determinism
- No implicit defaults.
- Images must be digest-pinned.
- All bindings are explicit in RunSpec.
- Derived state is computed from plan and step executions.

## Dry-run
- Deterministic simulation only.
- No user code execution.
- Records append-only step attempts.

## Derived run state
- Created: no plan.
- Planned: plan exists, no step executions.
- DryRunRunning: partial step outcomes.
- DryRunSucceeded: all steps succeeded or skipped.
- DryRunFailed: any step failed.

## Related docs
- `docs/pipeline-spec.md`
- `docs/adr/0003-execution-contracts-and-determinism.md`
