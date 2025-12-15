# Air-gapped install bundle

Animus DataPilot is designed to run without outbound network calls at runtime. For air-gapped installs, ship:

- Helm chart package (`deploy/helm/animus-datapilot`)
- Container images for all services + infra images (`postgres`, `minio`, `minio/mc`)

## Prerequisites

- Kubernetes cluster (on-prem or isolated)
- `helm` v3
- A container registry reachable by the cluster (or a way to preload images into the cluster runtime)

## Build / collect images (connected build environment)

Go services (repeat per service):

```bash
docker build -t animus/gateway:latest --build-arg SERVICE=gateway .
docker build -t animus/dataset-registry:latest --build-arg SERVICE=dataset-registry .
docker build -t animus/quality:latest --build-arg SERVICE=quality .
docker build -t animus/experiments:latest --build-arg SERVICE=experiments .
docker build -t animus/lineage:latest --build-arg SERVICE=lineage .
docker build -t animus/audit:latest --build-arg SERVICE=audit .
```

UI image (Next.js):

```bash
docker build -t animus/ui:latest -f frontend_landing/Dockerfile frontend_landing
```

Infra images (pull in a connected environment, then ship them into the air-gapped environment):

```bash
docker pull postgres:14-alpine
docker pull minio/minio@sha256:14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e
docker pull minio/mc@sha256:a7fe349ef4bd8521fb8497f55c6042871b2ae640607cf99d9bede5e9bdf11727
```

## Create an install bundle

The bundler does **not** download anything. It packages what is already in your local Docker image store.

```bash
./scripts/airgap-bundle.sh --output dist/airgap --image-repo animus --tag latest --include-infra 1
```

This produces:

- `animus-datapilot-*.tgz`
- `images.tar`
- `values.airgap.yaml`
- `SHA256SUMS`

## Install in the air-gapped environment

1) Verify integrity:

```bash
sha256sum -c SHA256SUMS
```

2) Load images on a host that can push to your in-cluster registry (or that can preload them into the cluster):

```bash
docker load -i images.tar
```

3) Push images to your internal registry as needed, and update `values.airgap.yaml` (`image.repository`, `minio.image`, etc).

4) Install:

```bash
helm upgrade --install animus-datapilot ./animus-datapilot-*.tgz -f values.airgap.yaml
```

## Notes

- The chart includes Postgres, MinIO, and a migration job by default. To use external infra, set `postgres.enabled=false` and/or `minio.enabled=false` and provide `database.url` / `minio.endpoint`.
- If using OIDC auth, serve UI under the same origin as the gateway (enable `ingress.enabled=true`); the ingress routes `/app` and `/_next` to the UI and `/` to the gateway.

