# 10. Versioning and Compatibility

## 10.1 Policy scope

Versioning governs all external contracts, including APIs, Pipeline Specification, evidence formats, and integration artifacts. Versioning is part of the architectural contract and is required for controlled upgrades.

## 10.2 Compatibility requirements

Compatibility requirements:

- Interfaces are explicitly versioned.
- Breaking changes require a version increment.
- Migration guidance is required for breaking changes.
- New features are additive where possible.
- Semantic meaning is preserved across versions.

## 10.3 Contracted artifacts

The following artifacts are subject to versioning and compatibility rules:

- API contracts and schemas.
- Pipeline Specification fields and constraints.
- Evidence bundle structure and metadata fields.
- CLI and SDK behavior as projections of the API contract.
- EnvironmentDefinition and EnvironmentLock formats.

## 10.4 Upgrade and rollback compatibility

Upgrades must preserve data integrity and maintain compatibility with persisted metadata.

Requirements:

- Schema migrations are controlled and reversible where possible.
- Rollback must not corrupt AuditEvent or domain history.
- Compatibility across upgrades is verified as part of operational readiness.

Operational procedures for upgrades and rollback are defined in Section 09.
