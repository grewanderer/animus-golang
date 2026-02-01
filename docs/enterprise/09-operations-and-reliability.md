# 09. Operations and Reliability

## 09.1 Operational objectives

Operations focus on predictable behavior under load, failure, and upgrades. Operational procedures must be documented, testable, and automatable.

## 09.2 Deployment model

Supported deployment environments:

- on-premise;
- private cloud;
- air-gapped installations.

Baseline execution environment for Data Plane is Kubernetes. Control Plane and Data Plane responsibilities remain separated in all deployment modes.

Delivery requirements:

- Helm and/or Kustomize are used for deployment artifacts.
- External database support is required for Control Plane metadata.
- Air-gapped installation is supported without external network dependencies.

## 09.3 Reliability requirements

Reliability requirements include:

- HA configuration for Control Plane components where required.
- Idempotent operations for create and update flows.
- Retry and backoff for transient failures.

Data Plane failures must not corrupt Control Plane metadata or audit history.

## 09.4 Observability

Observability requirements:

- Metrics for Control Plane, Data Plane, Run, PipelineRun, and scheduler queues.
- Structured logs collected centrally and scrubbed of secrets.
- Tracing across API requests, orchestration, and Control Plane to Data Plane interactions.

Observability data is part of operational acceptance (Section 12).

## 09.5 Backup and disaster recovery

Backup and DR requirements:

- Backup of metadata, audit records, and platform configuration.
- RPO and RTO targets defined per installation.
- Recovery procedures documented and validated.

## 09.6 Data governance

Governance requirements:

- Retention policies for DatasetVersion, Artifact, and AuditEvent.
- Secure deletion procedures.
- Legal hold support where required by policy.

## 09.7 Upgrades and migrations

Upgrades must be controlled and preserve data integrity.

Requirements:

- Updates are staged and support rollback.
- Schema migrations are controlled and reversible where possible.
- Breaking changes require explicit versioning and migration guidance.

Versioning policy is defined in Section 10.

## 09.8 CI/CD and quality gates

Platform delivery must include minimum quality gates:

- linting;
- unit tests;
- integration tests;
- end-to-end tests covering git to Run to Artifact to Model promotion;
- security scanning;
- SBOM generation.

## 09.9 Operational responsibilities

Operational responsibilities are defined for the following roles:

- Platform Owner: installation ownership and upgrade governance.
- SRE / Platform Engineer: reliability, monitoring, and incident response.
- Security Officer: security controls and compliance review.
- Project Maintainer: Project-level administration.

## 09.10 Operational runbooks

Operational runbooks are defined in [09-operational-runbooks.md](09-operational-runbooks.md).
