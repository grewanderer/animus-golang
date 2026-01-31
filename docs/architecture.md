# Architecture

Control, not a demo. Animus is a deterministic control plane for ML execution, lineage, and audit.

## Scope
- Control plane services manage metadata, policy, and audit.
- Data plane execution is external and pluggable.
- No user code runs inside the control plane.

## Deployment
- On-prem or private cloud.
- Air-gapped friendly; no external data flow required.
- Project scoping enforced at every layer.

## Controls
- Append-only audit and lineage records.
- Immutable execution contracts and deterministic planning.
- Explicit bindings for datasets, code, and environment.

## Outcome
- Reproducible runs with verifiable provenance.
- Evidence export suitable for audit and compliance.

## Planes

Control plane (closed/):
- Gateway and domain services (experiments, dataset registry, quality, lineage, audit).
- Persistence and audit ledger.
- Policy and validation.

Data plane (external):
- Runtime executor for user code.
- Managed by your platform team (Kubernetes, batch, CI runners).

## Storage and audit
- Project-scoped tables with strict filters.
- Append-only audit events with integrity checks.
- Evidence export for offline review.

## Network model
- All services run inside your network boundary.
- No outbound data flow is required for core operation.
- Integrations use explicit API contracts defined in `api/`.
