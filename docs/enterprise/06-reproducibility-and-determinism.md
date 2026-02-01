# 06. Reproducibility and Determinism

## 06.1 Scope and definition

Reproducibility is defined as the ability to re-execute a Run and obtain results that are consistent with the recorded inputs, policies, and execution environment.

Determinism is achieved only through explicit bindings and immutable references. The platform does not assume determinism when required bindings are missing or when non-deterministic behavior is introduced by user code or external dependencies.

## 06.2 Determinism inputs and bindings

A reproducible Run must bind and record the following inputs:

- DatasetVersion identifiers for all datasets consumed by the Run.
- CodeRef with commit SHA for the executed code.
- EnvironmentLock for the execution environment, including image digests and dependency checksums.
- Parameters and configuration values supplied to execution.
- Resource requirements and limits (CPU, RAM, GPU, storage).
- Applied policies and approvals.

Bindings are immutable once a Run is created.

## 06.3 Explicitness rules

Deterministic planning requires explicit declarations:

- Container images must be digest-pinned; tags are not sufficient for production-run.
- Environment variables must be fully declared; implicit inheritance is not allowed.
- Dependencies between steps must be explicit in the Pipeline Specification.
- Retry policy and resource requirements must be explicit for each execution step.
- No hidden defaults are allowed for fields that affect execution results.

## 06.4 Execution snapshot and replay

Control Plane must persist an immutable execution snapshot for each Run that includes all bindings in Section 06.2.

Replay is defined as the creation of a new Run using the same execution snapshot. Replay does not modify the original Run and must be recorded as a distinct Run with a reference to the source Run.

## 06.5 Production-run requirements

Production-run is defined in the glossary (Section 14). A production-run is accepted only if all of the following are true:

- CodeRef includes a commit SHA.
- EnvironmentLock is present and verifiable.
- DatasetVersion references are explicit and immutable.
- Policy approvals required by governance rules are satisfied and recorded.

If any requirement is unmet, the system must either reject the Run or record explicit limitations in the Run metadata and AuditEvent.

## 06.6 Dry-run and deterministic planning

Control Plane supports dry-run as a deterministic simulation that does not execute user code. Dry-run validates the Pipeline Specification, resolves bindings, and records step attempts without running Data Plane execution.

Derived run state is determined by the presence of an execution plan and step attempts, including:

- `created` (no plan exists);
- `planned` (plan exists, no step attempts);
- `dry_run_running` (partial step attempts);
- `dry_run_succeeded` (all steps succeeded or skipped);
- `dry_run_failed` (any step failed).

Dry-run results are auditable and must be recorded as part of the Run history.

## 06.7 Reproducibility and audit

Reproducibility claims are supported only when audit records and execution snapshots are complete. AuditEvent must capture determinism-relevant changes, including binding validation, policy application, and replay activity (see Section 05.8).

Any limitations to determinism must be recorded explicitly and linked to the Run for audit review.
