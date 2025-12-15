# ADR 0001: Pilot scope and architecture

## Status

Accepted

## Context

Animus DataPilot is an on-prem, enterprise-ready pilot focused on immutable dataset governance, quality gating, experiment tracking, lineage, and auditability.

Constraints:
- Must run without outbound network access at runtime.
- Must support on-prem identity providers.
- Must preserve an immutable history of data and metadata changes.
- Must provide a UI control plane that reads real state from backend APIs.

Explicitly out of scope:
- AutoML and training orchestration
- Annotation pipelines
- Streaming platforms (Kafka/Spark/Databricks)
- Multi-tenant SaaS billing
- External SaaS dependencies

## Decision

### Control plane

The UI is the single source of truth for the control plane. It reads and writes state exclusively through the gateway APIs and does not bypass service boundaries.

### Execution plane

Automation and CI integrations use an SDK and/or webhook endpoints to publish experiment/run metadata deterministically. Runtime behavior does not require external connectivity.

### Services

The pilot is composed of a gateway and domain services:
- `gateway`: authentication, authorization, routing, and cross-cutting concerns
- `dataset-registry`: dataset registration and immutable versioning
- `quality`: quality rule evaluation and enforcement
- `experiments`: experiment/run metadata and integrations
- `lineage`: lineage graph model and query APIs
- `audit`: centralized audit event ingestion and query APIs

### Persistence

Postgres is the system of record for domain metadata and audit events. MinIO is used for dataset and artifact object storage. Domain writes are append-only:
- Entities are immutable once created.
- New versions are expressed as new rows/records linked by stable identifiers.
- Corrections are new events/versions rather than in-place updates.

### Security

Authentication uses OIDC. Authorization uses RBAC enforced at the gateway and validated at service boundaries. All write actions emit auditable events.

## Consequences

- Immutability simplifies auditability and lineage but requires careful API design (create/version semantics instead of update-in-place).
- The gateway becomes a critical enforcement point; services still validate authorization to prevent confused deputy scenarios.
- On-prem constraints require packaging and local infrastructure for development and deployment (e.g., Postgres and MinIO).

