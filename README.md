![Animus — Enterprise ML Control Plane](docs/assets/animus-banner.png)

**Website:** — [kappaka.org](https://kapakka.org/en)
**Documentation:** `docs/`
**Deployment:** On-prem / Private cloud / Air-gapped

---

## Animus

Animus is an **enterprise ML control plane** for building, executing, and governing machine-learning workflows in a deterministic and auditable way.

Animus is designed for organizations that run ML **inside their own infrastructure** and require explicit control over data, execution, lineage, and decisions — without relying on hosted SaaS platforms or opaque automation.

Rather than acting as a training runtime or notebook environment, Animus serves as the **system of record and coordination layer** for ML development.

---

## What Animus Solves

In real production environments, ML systems fail not because of models, but because:

* datasets change without being noticed;
* experiments cannot be reproduced months later;
* pipelines rely on hidden state and ad-hoc scripts;
* audit evidence is reconstructed manually.

Animus addresses these problems by making the ML lifecycle **explicit, versioned, and traceable by default**.

---

## Key Capabilities

### Deterministic ML Execution

All experiments are bound to explicit inputs:

* immutable dataset versions;
* declared parameters;
* code commit identifiers;
* execution environment hashes.

If a result cannot be reproduced, Animus makes the reason visible.

---

### Declarative Pipelines

ML processes are described declaratively as pipelines that define *what* must happen, not *how* it is implemented.

This enables:

* predictable execution;
* reuse across teams;
* deterministic planning and dry-run simulation.

---

### Lineage and Audit by Design

Animus automatically records:

* dataset provenance and transformations;
* experiment parameters and outcomes;
* model approvals and lifecycle transitions;
* user actions and governance decisions.

All records are append-only and suitable for audit and review.

---

### Governance Without Friction

Quality checks, rules, and approval gates are defined declaratively and applied automatically, without requiring manual discipline from developers.

Governance becomes part of the workflow, not an external process.

---

## Architecture Overview

Animus is built around a strict separation of concerns.

### Control Plane

The control plane is responsible for:

* metadata and state management;
* orchestration and execution contracts;
* lineage, rules, and audit history.

The control plane **never executes user code**.

### Data Plane

The data plane is responsible for:

* pipeline execution;
* data processing and model training;
* artifact generation.

Execution is containerized and isolated. Kubernetes is the target runtime.

This separation ensures predictability, security, and operational stability.

See: `docs/architecture.md`

---

## Deployment Model

Animus is designed for enterprise environments:

* on-prem or private cloud deployment;
* support for air-gapped installations;
* no dependency on public SaaS services;
* no external data flow by default.

All services run **inside your network perimeter**.

---

## Getting Started

This repository includes an **open demo** that demonstrates the behavior of the control plane.

> The demo is not a hosted service and does not include a production data plane.

### Requirements

* Go 1.22+
* Docker
* Docker Compose
* `curl` (preferred) or `python3`

### Run the demo

```bash
make demo
```

### Smoke test

```bash
make demo-smoke
```

Stop with `Ctrl+C`. For cleanup: `make demo-down`.

---

## Repository Structure

```
.
├── open/       # Schemas, SDKs, demo runner, open documentation
├── closed/     # Control plane services, migrations, UI, deployment assets
├── api/        # Execution and API schemas
├── docs/       # Architecture, execution model, ADRs
```

---

## Documentation

* `docs/index.md` — documentation entry point
* `docs/architecture.md` — system architecture
* `docs/execution.md` — execution and determinism model
* `docs/pipeline-spec.md` — pipeline specification
* `docs/adr/` — architectural decision records

---

## Security

Animus is built for regulated and security-sensitive environments:

* SSO (OIDC / SAML)
* project-scoped RBAC
* secret isolation
* execution sandboxing
* full audit trail

For vulnerability reporting and coordinated disclosure, see `SECURITY.md`.

---

## Status

Animus is under active development.

This repository contains the **control plane foundation** of the Animus platform.

---

## License

See `LICENSE` for details.
