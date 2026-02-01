# README Traceability Matrix

This matrix maps each README section statement to `docs/enterprise/**` sources.

## Animus Datalab

- “Animus Datalab is an enterprise digital laboratory for machine learning that organizes the full ML development lifecycle in a managed, reproducible form within a single operational context with common execution, security, and audit rules.”
  - `docs/enterprise/01-system-definition-and-goals.md` §01.1 System definition

## Scope of this repository

- “`docs/enterprise/` contains the normative specification for Animus Datalab.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.1 Purpose and normative status (normative specification); repository path is a locator for those documents.
- “The normative specification defines system invariants, constraints, responsibilities, and acceptance criteria for Animus Datalab.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.1 Purpose and normative status
- “The specification describes the target production-grade state and is not tied to a specific implementation release.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.3 Document status and change control
- “The specification applies to on-premise, private cloud, and air-gapped deployments.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.4 Scope and applicability; `docs/enterprise/09-operations-and-reliability.md` §09.2 Deployment model

## System invariants

- “Control Plane does not execute user code.”
  - `docs/enterprise/01-system-definition-and-goals.md` §01.4 Architectural invariants; `docs/enterprise/03-architectural-model.md` §03.2 Control Plane; `docs/enterprise/14-glossary.md` Control Plane
- “A production-run is uniquely defined by DatasetVersion, CodeRef commit SHA, and EnvironmentLock.”
  - `docs/enterprise/01-system-definition-and-goals.md` §01.4 Architectural invariants; `docs/enterprise/06-reproducibility-and-determinism.md` §06.5 Production-run requirements
- “All significant actions are recorded as AuditEvent.”
  - `docs/enterprise/01-system-definition-and-goals.md` §01.4 Architectural invariants; `docs/enterprise/05-execution-model.md` §05.8 Execution model and audit; `docs/enterprise/08-security-model.md` §08.5 Audit
- “DatasetVersion, CodeRef commit SHA, EnvironmentLock, and Artifact are explicit entities recorded by the system.”
  - `docs/enterprise/02-conceptual-model.md` §02.3 Explicit ML experiment context; `docs/enterprise/04-domain-model.md` §04.3 DatasetVersion, §04.4 CodeRef, §04.5 EnvironmentLock, §04.7 Artifact
- “The system has no hidden state that affects execution results.”
  - `docs/enterprise/01-system-definition-and-goals.md` §01.4 Architectural invariants; `docs/enterprise/06-reproducibility-and-determinism.md` §06.3 Explicitness rules; `docs/enterprise/12-acceptance-criteria.md` AC-10
- “All domain entities belong to exactly one Project; cross-Project dependencies are prohibited.”
  - `docs/enterprise/04-domain-model.md` §04.2.3 Invariants; `docs/enterprise/04-domain-model.md` §04.1 Domain model principles

## Architectural model

- “Animus separates Control Plane and Data Plane responsibilities.”
  - `docs/enterprise/03-architectural-model.md` §03.1 Architecture overview
- “Control Plane manages governance, policies, metadata, orchestration, and AuditEvent generation and does not execute user code.”
  - `docs/enterprise/03-architectural-model.md` §03.1 Architecture overview; `docs/enterprise/03-architectural-model.md` §03.2 Control Plane
- “Data Plane executes user code in isolated, containerized environments and provides controlled access to DatasetVersion and Artifact.”
  - `docs/enterprise/03-architectural-model.md` §03.3 Data Plane; `docs/enterprise/05-execution-model.md` §05.5 Isolation and resource management; `docs/enterprise/08-security-model.md` §08.6 Execution isolation
- “Trust boundaries treat user clients as untrusted, Control Plane as trusted management, Data Plane as partially trusted execution, and external systems as separate zones integrated through contractual interfaces.”
  - `docs/enterprise/03-architectural-model.md` §03.4 Trust boundaries

## Domain model (summary)

- “Project is the isolation boundary and root container for DatasetVersion, Run, Artifact, ModelVersion, and AuditEvent.”
  - `docs/enterprise/04-domain-model.md` §04.2.2 Relationships; `docs/enterprise/04-domain-model.md` §04.2.3 Invariants; `docs/enterprise/14-glossary.md` Project
- “Run references DatasetVersion, CodeRef, and EnvironmentLock and produces Artifact.”
  - `docs/enterprise/02-conceptual-model.md` §02.2 Run as a unit of meaning; `docs/enterprise/04-domain-model.md` §04.6 Run; `docs/enterprise/04-domain-model.md` §04.7 Artifact
- “ModelVersion references a source Run and an Artifact.”
  - `docs/enterprise/04-domain-model.md` §04.8.3 Invariants; `docs/enterprise/14-glossary.md` ModelVersion
- “AuditEvent is append-only and spans the entity graph as the time dimension.”
  - `docs/enterprise/04-domain-model.md` §04.9 AuditEvent; `docs/enterprise/04-domain-model.md` §04.10 Domain model connectivity

## Execution and reproducibility

- “Run is the minimal unit of execution and reproducibility.”
  - `docs/enterprise/02-conceptual-model.md` §02.2 Run as a unit of meaning; `docs/enterprise/03-architecture-decision-records.md` ADR-002
- “A Run is defined by Project, DatasetVersion, CodeRef commit SHA, EnvironmentLock, execution parameters, and execution policy.”
  - `docs/enterprise/02-conceptual-model.md` §02.2 Run as a unit of meaning
- “Reproducibility is the ability to re-execute a Run and obtain results consistent with the recorded inputs, policies, and execution environment; determinism is not assumed when bindings are missing or user code introduces non-determinism.”
  - `docs/enterprise/06-reproducibility-and-determinism.md` §06.1 Scope and definition
- “Control Plane stores an immutable execution snapshot for each Run; replay creates a new Run from that snapshot and records the linkage.”
  - `docs/enterprise/06-reproducibility-and-determinism.md` §06.4 Execution snapshot and replay
- “A production-run is accepted only when CodeRef commit SHA, EnvironmentLock, DatasetVersion references, and required policy approvals are explicit and recorded; otherwise the Run is rejected or limitations are recorded in Run metadata and AuditEvent.”
  - `docs/enterprise/06-reproducibility-and-determinism.md` §06.5 Production-run requirements

## Security model (by design)

- “Authentication uses SSO via OIDC and/or SAML, with session TTL enforcement, forced logout, and limits on parallel sessions.”
  - `docs/enterprise/08-security-model.md` §08.2 Authentication
- “Authorization is Project-scoped with default deny; object-level constraints are enforced; service accounts are subject to the same RBAC and are audited.”
  - `docs/enterprise/08-security-model.md` §08.3 Authorization and RBAC; `docs/enterprise/08-rbac-matrix.md` §08.2.1 Scope; `docs/enterprise/08-rbac-matrix.md` §08.2.3 Service accounts
- “Secrets are supplied via an external secret store, are temporary and minimal in scope, are not exposed in UI, logs, metrics, or Artifact, and access attempts are recorded in AuditEvent.”
  - `docs/enterprise/08-security-model.md` §08.4 Secrets management
- “AuditEvent is append-only and cannot be disabled; audit export supports SIEM and monitoring integrations.”
  - `docs/enterprise/08-security-model.md` §08.5 Audit; `docs/enterprise/03-architecture-decision-records.md` ADR-006
- “Data Plane executes untrusted user code in containerized environments with restricted privileges; network access and resource limits are enforced by policy; Control Plane never executes user code.”
  - `docs/enterprise/08-security-model.md` §08.6 Execution isolation; `docs/enterprise/03-architectural-model.md` §03.3 Data Plane; `docs/enterprise/05-execution-model.md` §05.5 Isolation and resource management; `docs/enterprise/03-architectural-model.md` §03.2 Control Plane

## Explicit non-goals

- “Built-in Git hosting or full IDE replacement.”
  - `docs/enterprise/13-non-goals-and-exclusions.md`
- “A full inference platform (export to external systems is supported).”
  - `docs/enterprise/13-non-goals-and-exclusions.md`
- “A standalone Feature Store product (interfaces may be integrated).”
  - `docs/enterprise/13-non-goals-and-exclusions.md`

## Documentation

- “The authoritative specification is in `docs/enterprise/`.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.1 Purpose and normative status (normative specification); repository path is a locator for those documents.
- “This README is an entry point and is not the normative source.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.1 Purpose and normative status; `docs/enterprise/00-introduction-and-scope.md` §00.3 Document status and change control

## Status

- “The specification describes the target state and is not tied to a specific implementation release.”
  - `docs/enterprise/00-introduction-and-scope.md` §00.3 Document status and change control
- “Animus Datalab is production-grade only when all mandatory acceptance criteria are satisfied and verified on a working installation with security and audit policies enabled.”
  - `docs/enterprise/12-acceptance-criteria.md` §12.2 Production-grade definition
