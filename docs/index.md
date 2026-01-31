# Documentation

Control, not a demo. This documentation describes the deterministic control plane, on-prem deployment model, and execution contracts.

## Start here
- Architecture: `docs/architecture.md`
- Execution model: `docs/execution.md`
- Demo quickstart: `docs/demo.md`
- ADRs: `docs/adr/README.md`

## Navigation
- Architecture: `docs/architecture.md`
- Execution contracts: `docs/execution.md`
- PipelineSpec schema guide: `docs/pipeline-spec.md`
- Demo quickstart: `docs/demo.md`
- Style guide: `docs/style.md`
- ADRs: `docs/adr/README.md`
- Kubernetes baseline (runtime posture): `docs/kubernetes-baseline.md`

## Determinism
- Execution is specified before it is run.
- RunSpec binds dataset versions, commit SHA, and env hash.
- Derived state comes from plan presence and step executions.

## Security
- Project isolation is mandatory.
- Audit events are append-only and non-optional.
- Control plane does not execute user code.

## Operability
- Idempotent run creation.
- Deterministic dry-run simulation.
- Exportable evidence and audit logs.
