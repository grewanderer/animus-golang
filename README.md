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
Requirements: Go, Docker, docker compose (or docker-compose), curl (preferred) or python3.

One command:

```bash
make demo
```

CI-style smoke check:

```bash
make demo-smoke
```

Expected output (example):

```
==> starting demo stack
==> waiting for gateway http://localhost:8080/healthz
==> create project
==> create run
==> dry-run
==> userspace execution (data plane surface)
==> demo complete
```

Stop: Ctrl+C (the script shuts down services on exit). For manual cleanup: `make demo-down`.

Troubleshooting:
- Port already in use: set `ANIMUS_GATEWAY_PORT`.
- Missing curl: install curl or python3.
- Go toolchain mismatch: install Go 1.22+.

Docs locations:
- Open demo docs live in `open/demo/`.
- Platform docs live in `docs/`.

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
