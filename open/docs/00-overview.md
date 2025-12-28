# Overview

Animus DataPilot is an on-prem, air-gapped compatible control plane for datasets, experiments, lineage, and auditability. It enforces immutability and produces audit evidence for every write operation.

## Repository scope

This repository contains the **open integration layer**: OpenAPI specifications, SDKs, demo clients, and documentation. The control-plane services, UI, and deployment assets are closed-source and are not included here.

## Problem

Enterprise ML teams need deterministic, auditable runs that survive audits and security reviews. Traditional tooling often depends on outbound SaaS or mutable metadata, which is unacceptable in regulated or air-gapped environments.

## Solution

Animus DataPilot provides:

- Immutable dataset versions and experiment runs.
- Explicit quality gates before downstream use.
- Lineage edges that connect dataset → run → git commit → image digest.
- Audit logs and evidence bundles for compliance review.
- A UI control plane (closed) and a minimal SDK (open) for CI and training containers.

## Value propositions

- Deterministic execution records tied to dataset hash, git commit, and image digest.
- Audit-ready evidence bundles (ledger, lineage, audit slice, policy snapshot, report).
- On-prem first deployment with no outbound dependencies for runtime operation.
- Clear security boundaries (gateway-only access, signed internal headers, run tokens).

## Personas

- ML engineers: register datasets, run experiments, log metrics, and collect artifacts.
- Platform engineers: deploy on-prem, enforce RBAC, manage upgrades and backups.
- Security and compliance: review evidence bundles, audit trails, and lineage.
- SRE/DevOps: monitor services and manage day-2 operations.

## What this is

- A deterministic control plane for ML metadata and execution evidence.
- An open integration kit (schemas, SDKs, and demos) that targets the closed-core services.

## What this is not

- AutoML or training orchestration beyond container execution.
- A notebook or labeling platform.
- A multi-tenant SaaS billing system.

## Related docs

- `01-architecture.md`
- `02-security-and-compliance.md`
- `03-deployment.md`
- `06-cli-and-usage.md`
- `10-glossary.md`
