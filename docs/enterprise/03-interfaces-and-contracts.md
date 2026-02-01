# 03.6 Interfaces and Contracts

## 03.6.1 Interface principles

Interfaces are part of the architectural contract and follow these principles:

- API is the source of truth for all operations on entities.
- UI does not use hidden or unstable interfaces.
- CLI and SDK are built on top of the public API.
- Interfaces are versioned; compatibility and breaking changes are governed by the versioning policy (Section 10).
- Critical operations must be idempotent and repeatable.

## 03.6.2 API resource model

The API follows a resource model in which each domain entity is an addressable resource with an explicit lifecycle.

Baseline API characteristics:

- implementation may use REST or gRPC while preserving the contract;
- strict authentication and authorization;
- explicit, machine-readable errors;
- correlation/request ID support.

Canonical resource set includes:

- Project;
- Dataset and DatasetVersion;
- EnvironmentDefinition and EnvironmentLock;
- Run and PipelineRun;
- Artifact;
- Model and ModelVersion;
- AuditEvent.

Each resource must support:

- retrieval (GET);
- creation (POST);
- filtering and search within access rights;
- change history where applicable.

## 03.6.3 Operation semantics

Create:

- creation must be atomic;
- invariant violations return errors;
- successful creation is recorded in AuditEvent.

Read:

- data is returned strictly within access rights;
- missing access returns an explicit reasoned error.

Update:

- permitted only for entities with mutable state;
- immutable entities do not allow updates.

Delete:

- physical deletion is restricted by retention and legal hold policies;
- logical deletion (archive/deprecate) is preferred.

## 03.6.4 Errors and diagnostics

API errors must be:

- typed;
- machine-readable;
- accompanied by diagnostic context.

Minimum error fields:

- `error_code`;
- `message`;
- `details`;
- `request_id`.

## 03.6.5 Pipeline Specification

Pipeline Specification (see Section 14) describes the declarative structure of ML execution and serves as the orchestration contract. Pipeline Specification does not contain executable code.

Pipeline Specification must include:

- identifier and specification version;
- list of steps (nodes);
- dependencies between steps;
- step inputs and outputs;
- error and retry policies;
- resource requirements.

Before creating a PipelineRun, Control Plane must:

- validate DAG correctness (no cycles);
- validate all references;
- validate against security policies;
- reject the specification on invariant violations.

## 03.6.6 Events and integrations

Events reflect state changes and are not a source of truth.

Minimum required event set:

- `DatasetVersionCreated`;
- `RunStarted`;
- `RunFinished`;
- `PipelineRunFinished`;
- `ModelVersionCreated`;
- `ModelApproved`;
- `AuditEventExported`.

Each event must:

- include a reference to the object;
- include a timestamp;
- be safely redeliverable without breaking consistency.

Supported delivery mechanisms:

- webhook;
- message broker;
- export to SIEM/observability systems.

The system must account for redelivery and temporary receiver outages.

## 03.6.7 CLI and SDK

CLI and SDK are used for automation, CI/CD integration, and non-UI usage.

Requirements:

- CLI and SDK use the API and do not bypass Control Plane policies;
- authentication via SSO or service accounts is supported;
- CLI is scriptable and returns machine-readable results;
- SDK mirrors the API resource model and does not hide errors or platform constraints.

## 03.6.8 UI

UI is a visual representation of platform state and is not a logic source.

UI must not:

- perform actions not available through the API;
- hide information required for audit or diagnostics.

## 03.6.9 Interfaces, security, and audit

Every interaction via interfaces must:

- be authenticated;
- be authorized against access policies;
- be recorded in audit;
- be traceable via request/correlation ID.
