# Animus DataPilot

<p align="center">
  <strong>Control plane for auditable, on-prem ML execution and lineage</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-production--ready-0f172a?style=flat-square" />
  <img src="https://img.shields.io/badge/deployment-on--prem--first-0f172a?style=flat-square" />
  <img src="https://img.shields.io/badge/network-air--gapped--compatible-0f172a?style=flat-square" />
  <img src="https://img.shields.io/badge/model-open--core-0f172a?style=flat-square" />
  <img src="https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square" />
</p>

---

## Overview

**Animus DataPilot** is an enterprise-grade control plane for governed machine-learning execution. It enables organizations to *prove* how models were trained by enforcing deterministic lineage, policy checks, and immutable audit records.

The system is designed for environments where ML workflows must be explainable, reviewable, and defensible — including regulated, on-prem, and air-gapped deployments.

Animus follows an **open-core architecture**:
- open, stable integration interfaces (schemas, SDKs, documentation)
- commercial control-plane implementation for production use

---

## Why Animus exists

Most ML failures in regulated environments are not caused by model quality, but by missing or unverifiable execution history.

Common failure modes include:

- datasets changing without traceability
- pipeline behavior drifting over time
- code revisions not tied to executions
- implicit approvals and undocumented actions
- audit evidence reconstructed manually after the fact

Animus exists to make ML execution **provable, deterministic, and auditable by design**.

---

## Design principles

### Immutability by default
All datasets, lineage edges, and audit records are append-only and never mutated.

### Audit is mandatory
Every state-changing operation produces a verifiable audit event.

### Explicit source of intent
User intent is expressed via the control interface; authorization, validation, and persistence are enforced by backend services.

### CI as the execution plane
SDKs and integrations are designed for non-interactive execution in CI/CD systems.

### On-prem first
Cloud deployment is optional. Air-gapped environments are fully supported.

---

## High-level architecture

## Trust and verification model

Animus is designed so that claims about ML execution can be independently verified.

Key properties:

- identifiers are content-derived or cryptographically bound
- lineage edges are immutable
- audit records are append-only
- evidence can be exported and verified offline
- no hidden or implicit state exists outside the control plane

This allows auditors, security teams, or third parties to validate execution history without trusting runtime operators.

---

## Typical integration flow

1. Register datasets and schemas
2. Define quality and policy constraints
3. Execute training or evaluation in CI
4. Report execution metadata to Animus
5. Generate immutable lineage and audit records
6. Export evidence for review or compliance

All steps can be executed in fully on-prem or air-gapped environments.

---

## Open vs commercial components

This repository contains the **open integration surface** used by client systems and CI pipelines.

### Included (open)

- OpenAPI specifications
- SDKs for CI integration
- Demo tooling and reference workflows
- Integration and operator documentation

### Commercial (not included)

- Control-plane services
- Policy enforcement engine
- Web UI
- Audit and evidence storage backend
- Identity, RBAC, and SSO integrations
- Deployment automation and upgrades

All commercial components operate strictly on top of the open interfaces defined here. No proprietary APIs are required to integrate with Animus.

---

## Repository structure

- `closed/` — commercial services, UI, deploy, migrations, e2e
- `docs/` — specs and technical requirements
- `tools/` — CI helpers and scripts

---

## Getting started

Run repository checks:

```bash
make lint
make test
make build
```

---

## Development

Format, lint, test, and build:

```bash
make fmt
make lint
make test
make build
```

Install `golangci-lint` if missing:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Documentation

- [`docs/Итоговое техническое задание.md`](docs/Итоговое техническое задание.md)
- Execution and API schemas are located under the top-level `api/` directory.

### Artifact endpoints (dataset registry)

Artifacts are registered in the control plane and transferred via short-lived pre-signed URLs (default 10 minutes).

- `POST /projects/{project_id}/artifacts` → returns `upload_url` (direct PUT to object store)
- `GET /projects/{project_id}/artifacts/{artifact_id}` → metadata
- `GET /projects/{project_id}/artifacts/{artifact_id}/download` → returns `download_url` (direct GET from object store)
- [`docs/kubernetes-baseline.md`](docs/kubernetes-baseline.md)
- [`docs/adr/0001-control-plane-data-plane.md`](docs/adr/0001-control-plane-data-plane.md)
- [`docs/adr/0002-project-isolation.md`](docs/adr/0002-project-isolation.md)

---

## Security

For vulnerability reporting, security questions, or coordinated disclosure, see [`SECURITY.md`](SECURITY.md).

---

## Intended audience

- ML platform and infrastructure engineers
- Security and compliance teams
- Regulated organizations deploying ML
- Internal AI enablement teams
