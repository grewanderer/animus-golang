````md
<p align="center">
  <img src="assets/banner.png" width="100%" alt="Animus Memory Core">
</p>

<h1 align="center">Animus DataPilot</h1>
<p align="center">
  <em>Deterministic control plane for enterprise ML datasets, experiments, and lineage</em>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-production--ready-0f172a?style=flat-square">
  <img src="https://img.shields.io/badge/deployment-on--prem--first-0f172a?style=flat-square">
  <img src="https://img.shields.io/badge/network-air--gapped--compatible-0f172a?style=flat-square">
  <img src="https://img.shields.io/badge/go-1.22+-0f172a?style=flat-square">
</p>

---

## What is this?

**Animus DataPilot** is an **enterprise-grade control plane** for managing:

- datasets and immutable dataset versions
- quality gates and enforcement
- experiments and CI-bound runs
- full lineage (dataset → run → git commit)
- auditable actions across the entire system

This is **not**:
- AutoML
- a notebook platform
- a SaaS-first product
- a data labeling system

This **is**:
- deterministic
- auditable
- on-prem ready
- air-gapped compatible

---

## Design intent

> Large organizations do not want magic.
> They want control, immutability, and answers.

Animus DataPilot is designed for environments where:
- outbound network access is restricted or forbidden
- data must never leave the perimeter
- every action must be explainable months later
- CI/CD is the execution plane
- UI is the control plane

---

## High-level architecture

```text
┌──────────────────────────────┐
│             UI               │  Control plane
│   (Next.js + Tailwind)       │
└──────────────┬───────────────┘
               │
┌──────────────▼───────────────┐
│           Gateway            │  Auth · RBAC · Audit
└──────────────┬───────────────┘
               │
┌──────────────▼───────────────┐
│  Core Services (Go)          │
│  ─ Dataset Registry          │
│  ─ Quality                   │
│  ─ Experiments               │
│  ─ Lineage                   │
│  ─ Audit                     │
└──────────────┬───────────────┘
               │
┌──────────────▼───────────────┐
│ Infrastructure               │
│ ─ Postgres (metadata)        │
│ ─ MinIO (artifacts)          │
└──────────────────────────────┘
````

Everything is explicit.
Nothing is inferred.

---

## Core principles

* **Immutability by default**
  Datasets, versions, experiments, lineage edges are never mutated.

* **Audit is not optional**
  Every write operation produces an audit event.

* **UI is the single source of truth**
  No hidden state in SDKs or clients.

* **CI is the execution plane**
  SDKs are designed for pipelines, not notebooks.

* **On-prem first**
  Cloud is optional. Air-gapped is supported.

---

## What is implemented

### Services

| Service            | Responsibility                            |
| ------------------ | ----------------------------------------- |
| `gateway`          | Auth, RBAC, request validation, audit     |
| `dataset-registry` | Datasets and immutable versions           |
| `quality`          | Declarative quality rules and enforcement |
| `experiments`      | Experiments, runs, CI metadata            |
| `lineage`          | Dataset → run → git lineage graph         |
| `audit`            | Central audit log                         |
| `ui`               | Control plane UI                          |

### SDKs

* **Python SDK**

  * Register experiment runs
  * Attach metrics
  * Bind Git metadata
  * CI-safe, no outbound calls

### Deployment

* `docker-compose` for local development
* Helm chart for Kubernetes
* Air-gapped installation bundle
* Deterministic demo scenario included

---

## Repository structure

```text
.
├── gateway/
├── dataset-registry/
├── quality/
├── experiments/
├── lineage/
├── audit/
├── sdk/
│   └── python/
├── frontend_landing/
├── internal/
│   └── platform/
├── deploy/
│   └── helm/
├── migrations/
├── docs/
├── demo/
└── Makefile
```

---

## Getting started (local)

```bash
make bootstrap
make dev
make migrate-up
```

Run full verification:

```bash
make lint
make test
make e2e
```

---

## Definition of Done (enforced)

* no TODOs
* no stub implementations
* no fake data outside demo runner
* every service compiles and runs
* every API documented
* UI renders real backend data

If it does not pass `make test`, it does not exist.

---

## Documentation

* [`docs/architecture.md`](docs/architecture.md)
* [`docs/auth.md`](docs/auth.md)
* [`docs/quality.md`](docs/quality.md)
* [`docs/experiments.md`](docs/experiments.md)
* [`docs/lineage.md`](docs/lineage.md)
* [`docs/audit.md`](docs/audit.md)
* [`docs/deploy.md`](docs/deploy.md)
* [`docs/airgap.md`](docs/airgap.md)
* [`docs/demo.md`](docs/demo.md)

---

## Intended audience

* Enterprise ML platform teams
* Regulated industries
* Internal AI/ML enablement groups
* Organizations tired of opaque pipelines
