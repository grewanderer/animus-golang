![Animus Datalab](docs/assets/animus-banner.png)

Animus Datalab is an enterprise digital laboratory for machine learning that organizes the full ML development lifecycle in a managed, reproducible form within a single operational context with common execution, security, and audit rules.

## Scope of this repository

- `docs/enterprise/` contains the normative specification for Animus Datalab.
- The normative specification defines system invariants, constraints, responsibilities, and acceptance criteria for Animus Datalab.
- The specification describes the target production-grade state and is not tied to a specific implementation release.
- The specification applies to on-premise, private cloud, and air-gapped deployments.

## System invariants

- Control Plane does not execute user code.
- A production-run is uniquely defined by DatasetVersion, CodeRef commit SHA, and EnvironmentLock.
- All significant actions are recorded as AuditEvent.
- DatasetVersion, CodeRef commit SHA, EnvironmentLock, and Artifact are explicit entities recorded by the system.
- The system has no hidden state that affects execution results.
- All domain entities belong to exactly one Project; cross-Project dependencies are prohibited.

## Architectural model

Animus separates Control Plane and Data Plane responsibilities. Control Plane manages governance, policies, metadata, orchestration, and AuditEvent generation and does not execute user code. Data Plane executes user code in isolated, containerized environments and provides controlled access to DatasetVersion and Artifact. Trust boundaries treat user clients as untrusted, Control Plane as trusted management, Data Plane as partially trusted execution, and external systems as separate zones integrated through contractual interfaces.

## Domain model (summary)

- Project is the isolation boundary and root container for DatasetVersion, Run, Artifact, ModelVersion, and AuditEvent.
- Run references DatasetVersion, CodeRef, and EnvironmentLock and produces Artifact.
- ModelVersion references a source Run and an Artifact.
- AuditEvent is append-only and spans the entity graph as the time dimension.

## Execution and reproducibility

Run is the minimal unit of execution and reproducibility. A Run is defined by Project, DatasetVersion, CodeRef commit SHA, EnvironmentLock, execution parameters, and execution policy. Reproducibility is the ability to re-execute a Run and obtain results consistent with the recorded inputs, policies, and execution environment; determinism is not assumed when bindings are missing or user code introduces non-determinism. Control Plane stores an immutable execution snapshot for each Run; replay creates a new Run from that snapshot and records the linkage. A production-run is accepted only when CodeRef commit SHA, EnvironmentLock, DatasetVersion references, and required policy approvals are explicit and recorded; otherwise the Run is rejected or limitations are recorded in Run metadata and AuditEvent.

## Security model (by design)

- Authentication uses SSO via OIDC and/or SAML, with session TTL enforcement, forced logout, and limits on parallel sessions.
- Authorization is Project-scoped with default deny; object-level constraints are enforced; service accounts are subject to the same RBAC and are audited.
- Secrets are supplied via an external secret store, are temporary and minimal in scope, are not exposed in UI, logs, metrics, or Artifact, and access attempts are recorded in AuditEvent.
- AuditEvent is append-only and cannot be disabled; audit export supports SIEM and monitoring integrations.
- Data Plane executes untrusted user code in containerized environments with restricted privileges; network access and resource limits are enforced by policy; Control Plane never executes user code.

## Explicit non-goals

- Built-in Git hosting or full IDE replacement.
- A full inference platform (export to external systems is supported).
- A standalone Feature Store product (interfaces may be integrated).

## Documentation

The authoritative specification is in `docs/enterprise/`. This README is an entry point and is not the normative source.

## Status

The specification describes the target state and is not tied to a specific implementation release. Animus Datalab is production-grade only when all mandatory acceptance criteria are satisfied and verified on a working installation with security and audit policies enabled.
