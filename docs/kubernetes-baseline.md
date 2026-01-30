# Kubernetes Baseline

Date: 2026-01-30

This repository targets Kubernetes as the default runtime baseline for production
deployments. The baseline supports:

- single-cluster deployments
- multi-cluster readiness (control plane to data plane separation)
- air-gapped environments

## Current deployment assets
- Helm chart: `closed/deploy/helm/animus-datapilot`
- Docker Compose (local/dev): `closed/deploy/docker-compose.yml`

## Notes
- Control Plane services run as standard deployments.
- Data Plane execution uses Kubernetes Jobs created by the runtime executor.
- Air-gapped installs should rely on pre-pulled images and internal registries.
