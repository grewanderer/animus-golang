# ADR 0002: Project Isolation Boundaries

Date: 2026-01-30
Status: Accepted

## Context
The specification requires Projects to be hard isolation boundaries with no hidden
cross-project dependencies.

## Decision
- Every primary entity (dataset, run, artifact, model, audit event) is scoped to a
  Project identifier at the storage and API layers.
- Cross-project references are disallowed by default and must be explicit if ever
  introduced in future extensions.
- Any data plane execution must be bound to the originating Project context.

## Consequences
- Repository interfaces will require Project identifiers for reads/writes.
- Governance and access checks can enforce project-level RBAC consistently.
- Migration and schema changes must preserve project scoping.
