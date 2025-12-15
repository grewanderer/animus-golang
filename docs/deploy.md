# Deployment (enterprise)

This pilot is designed to run on-prem (including air-gapped environments) and integrate with enterprise identity via OIDC.

## Build and publish images

Build Go services:

```bash
docker build -t <REGISTRY>/animus/gateway:<TAG> --build-arg SERVICE=gateway .
docker build -t <REGISTRY>/animus/dataset-registry:<TAG> --build-arg SERVICE=dataset-registry .
docker build -t <REGISTRY>/animus/quality:<TAG> --build-arg SERVICE=quality .
docker build -t <REGISTRY>/animus/experiments:<TAG> --build-arg SERVICE=experiments .
docker build -t <REGISTRY>/animus/lineage:<TAG> --build-arg SERVICE=lineage .
docker build -t <REGISTRY>/animus/audit:<TAG> --build-arg SERVICE=audit .
```

Build UI:

```bash
docker build -t <REGISTRY>/animus/ui:<TAG> -f frontend_landing/Dockerfile frontend_landing
```

Push to your internal registry:

```bash
docker push <REGISTRY>/animus/gateway:<TAG>
# ...repeat for all images...
```

## Install with Helm

1) Configure values (start from `deploy/helm/animus-datapilot/values.yaml`):

Key values:

- `image.repository` + `image.tag`: your registry and tag
- `auth.internalAuthSecret`: shared HMAC for gateway→services request signing
- `ci.webhookSecret`: HMAC for CI webhook ingestion (`/api/experiments/ci/webhook`)
- `database.url` and `postgres.enabled` (use in-cluster Postgres or external)
- `minio.enabled` and `minio.endpoint` (use in-cluster MinIO or external S3-compatible storage)

OIDC (recommended for enterprise):

- `auth.mode: oidc`
- `oidc.issuerURL`, `oidc.clientID`
- `oidc.clientSecret` + `oidc.redirectURL` (required to enable gateway `/auth/login` flow)
- `auth.sessionCookieSecure: true` when served over HTTPS

2) Install:

```bash
helm upgrade --install animus-datapilot deploy/helm/animus-datapilot -f values.enterprise.yaml
```

3) Expose:

- Use your ingress controller by setting `ingress.enabled=true` and `ingress.host`.
- The chart routes `/app` and `/_next` to the UI and `/` to the gateway so OIDC cookies work under one origin.

## Recommended enterprise hardening

- Network: only expose `gateway` (+ `ui`); keep other services internal and enforce NetworkPolicies.
- Secrets: store `auth.internalAuthSecret`, `ci.webhookSecret`, and `oidc.clientSecret` in a proper secret manager (sealed-secrets/external-secrets).
- TLS: terminate HTTPS at ingress; set `auth.sessionCookieSecure=true`.
- Backups: back up Postgres + object storage buckets (`datasets`, `artifacts`).
- Observability: ship JSON logs from all pods to your SIEM/log platform; correlate by `X-Request-Id`.

