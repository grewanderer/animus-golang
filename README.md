# Animus

Control, not a demo. Deterministic, auditable ML execution for on-prem and private cloud.

## Scope
- Control plane for execution contracts, lineage, and audit.
- No user code execution in the control plane.
- Project-scoped APIs and storage.

## Deployment
- On-prem or private cloud.
- Air-gapped friendly; no external data flow.
- Services run inside your network boundary.

## Controls
- Immutable lineage and append-only audit.
- Deterministic execution contracts (PipelineSpec, RunSpec, ExecutionPlan).
- Idempotent run creation and dry-run simulation.

## Outcome
- Reproducible runs bound to dataset version, commit SHA, and environment hash.
- Verifiable evidence export without external dependencies.

## Guarantees
- Determinism: explicit bindings, no hidden defaults.
- Security: project isolation, audit-on-write, on-prem data plane.
- Operability: append-only records, idempotent APIs, exportable evidence.

## Quickstart (open demo)
Requirements: Go, Docker, docker compose, curl.

One command:

```bash
make demo
```

Expected output includes:
- "==> services are running"
- "==> animus demo (gateway=...)"
- "==> smoke check ok"

Troubleshooting:
- Port 8080 in use: set `ANIMUS_GATEWAY_PORT`.
- Docker not running: start Docker or Docker Desktop.
- Missing Go toolchain: install Go 1.22+.

## Architecture
- Control plane services: closed/ (experiments, dataset-registry, quality, lineage, audit).
- Data plane execution: external runtime, not in this repo.
- Gateway: API entry point with project scoping.

See `docs/architecture.md`.

## Execution contracts
- PipelineSpec: declarative template with datasetRef.
- RunSpec: binds datasetRef -> datasetVersionId, commit SHA, env hash.
- ExecutionPlan: deterministic DAG ordering.
- Dry-run: deterministic simulation, no user code execution.
- Derived run state: computed from plan presence and step executions.

See `docs/execution.md` and `docs/pipeline-spec.md`.

## Repo map
- `open/`: schemas, SDKs, demo runner, open docs.
- `closed/`: control plane services, migrations, deployment assets, UI.
- `api/`: execution and API schemas.
- `docs/`: architecture and execution documentation.

## Documentation
- `docs/index.md`
- `docs/architecture.md`
- `docs/execution.md`
- `docs/demo.md`
- `docs/style.md`
- `docs/adr/README.md`

## Security
For vulnerability reporting or coordinated disclosure, see `SECURITY.md`.
