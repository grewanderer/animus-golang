# Kubernetes baseline

Date: 2026-01-30

## Scope
- Kubernetes is the baseline runtime target for data plane execution.
- Control plane services remain separate and do not run user code.

## Deployment
- On-prem or private cloud clusters.
- Air-gapped compatible with internal registries.

## Controls
- Project scoping enforced by the control plane.
- Execution contracts validated before any runtime integration.

## Outcome
- Runtime selection stays pluggable.
- Kubernetes integration is a deployment decision, not a hard-coded dependency.

## Current assets
- Helm chart: `closed/deploy/helm/animus-datapilot`
- Docker Compose (local/dev): `closed/deploy/docker-compose.yml`

## Notes
- The open demo and dry-run executor do not run user code.
- Runtime executors are integrated later via the data plane.
