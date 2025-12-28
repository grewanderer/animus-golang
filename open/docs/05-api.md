# API Guide

This guide summarizes the API surface and usage patterns exposed by the Animus gateway (closed-core). The canonical definitions are the OpenAPI specs in `open/api/openapi/`.

## Base URLs and routing

- Gateway base URL: `http://localhost:8080`
- Gateway proxies service APIs under:
  - `/api/dataset-registry/*`
  - `/api/quality/*`
  - `/api/experiments/*`
  - `/api/lineage/*`
  - `/api/audit/*`

## Authentication

### OIDC (default)

- Use `Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9` for API calls.
- For browser sessions, the gateway sets an auth cookie via `/auth/login` and `/auth/callback`.

### Dev mode

- `AUTH_MODE=dev` authenticates every request as the configured dev identity; no token required.

### Disabled mode

- `AUTH_MODE=disabled` accepts all requests without authentication (development only).

### Run tokens (training containers)

Training/evaluation containers receive a run-scoped bearer token and must call the gateway with:

```
Authorization: Bearer animus_run_v1.eyJydW5faWQiOiJydW5fMSIsImV4cCI6MTk5OTk5OTk5OX0.QWJjMTIz
```

Run tokens are restricted to the run ID and (if provided) dataset version ID.

## Error model

Services return JSON errors with a stable shape:

```json
{
  "error": "quality_gate_failed",
  "request_id": "req_01J1X9K7B3ZJ4A1XH6Y1C9QZ8Q"
}
```

Some endpoints include `details` for validation errors (e.g., dataset upload size limits).

## Pagination

- Most list endpoints accept `limit` with max values (typically 500).
- Run events and lineage/audit lists support cursor pagination via `before_event_id`.
- SSE stream supports `after_event_id` to resume.

## Idempotency and rate limits

- Idempotency keys are not implemented in this repository.
- Rate limiting is not implemented; enforce at ingress or API gateway.

## Endpoint summary

### Gateway

- `GET /healthz`
- `GET /readyz`
- `GET /auth/session`
- `POST /auth/logout`
- `GET /auth/login`
- `GET /auth/callback`
- Proxy routes under `/api/*` (see services below)

### Dataset Registry (`/api/dataset-registry`)

- `GET /datasets`
- `POST /datasets`
- `GET /datasets/{dataset_id}`
- `GET /datasets/{dataset_id}/versions`
- `POST /datasets/{dataset_id}/versions/upload`
- `GET /dataset-versions/{version_id}`
- `GET /dataset-versions/{version_id}/download`

### Quality (`/api/quality`)

- `GET /rules`
- `POST /rules`
- `GET /rules/{rule_id}`
- `GET /evaluations?dataset_version_id=dv_example_01`
- `POST /evaluations`
- `GET /evaluations/{evaluation_id}`
- `GET /gates/dataset-versions/{version_id}`

### Experiments (`/api/experiments`)

- `GET /experiments`
- `POST /experiments`
- `GET /experiments/{experiment_id}`
- `GET /experiments/{experiment_id}/runs`
- `POST /experiments/{experiment_id}/runs`
- `POST /experiments/runs:execute`
- `GET /experiment-runs`
- `GET /experiment-runs/{run_id}`
- `GET /experiment-runs/{run_id}/execution`
- `GET /experiment-runs/{run_id}/build-context`
- `GET /experiment-runs/{run_id}/metrics`
- `POST /experiment-runs/{run_id}/metrics`
- `GET /experiment-runs/{run_id}/artifacts`
- `POST /experiment-runs/{run_id}/artifacts`
- `GET /experiment-runs/{run_id}/artifacts/{artifact_id}`
- `GET /experiment-runs/{run_id}/artifacts/{artifact_id}/download`
- `GET /experiment-runs/{run_id}/evidence-bundles`
- `POST /experiment-runs/{run_id}/evidence-bundles`
- `GET /experiment-runs/{run_id}/evidence-bundles/{bundle_id}`
- `GET /experiment-runs/{run_id}/evidence-bundles/{bundle_id}/download`
- `GET /experiment-runs/{run_id}/evidence-bundles/{bundle_id}/report`
- `GET /experiment-runs/{run_id}/stream`
- `GET /experiment-runs/{run_id}/events`
- `POST /experiment-runs/{run_id}/events`
- `GET /execution-ledger`
- `GET /execution-ledger/{run_id}`
- `POST /ci/webhook`
- `POST /ci/report`
- `POST /gitlab/webhook`
- `GET /model-images`
- `GET /model-images/{image_digest}`
- `GET /policies`
- `POST /policies`
- `GET /policies/{policy_id}`
- `GET /policies/{policy_id}/versions`
- `POST /policies/{policy_id}/versions`
- `GET /policy-decisions`
- `GET /policy-decisions/{decision_id}`
- `GET /policy-approvals`
- `GET /policy-approvals/{approval_id}`
- `POST /policy-approvals/{approval_id}/approve`
- `POST /policy-approvals/{approval_id}/deny`

CI webhook endpoints require signature headers:

- `X-Animus-CI-Ts` (Unix timestamp)
- `X-Animus-CI-Sig` (HMAC-SHA256 over `ts\\nMETHOD\\nsha256(body)` with `ANIMUS_CI_WEBHOOK_SECRET`)

### Lineage (`/api/lineage`)

- `GET /events`
- `GET /subgraphs/datasets/{dataset_id}`
- `GET /subgraphs/dataset-versions/{version_id}`
- `GET /subgraphs/experiment-runs/{run_id}`
- `GET /subgraphs/git-commits/{commit}`

### Audit (`/api/audit`)

- `GET /events`
- `GET /events/{event_id}`

## Example requests

### Create a dataset

```bash
curl -sS -X POST http://localhost:8080/api/dataset-registry/datasets \
  -H 'Content-Type: application/json' \
  -d '{"name":"fraud-dataset","description":"Fraud training set","metadata":{"owner":"ml-team"}}'
```

Example response:

```json
{
  "dataset_id": "ds_8b2e7db2",
  "name": "fraud-dataset",
  "description": "Fraud training set",
  "metadata": {"owner": "ml-team"},
  "created_at": "2025-01-10T12:00:00Z",
  "created_by": "dev-user"
}
```

### Upload a dataset version

```bash
curl -sS -X POST http://localhost:8080/api/dataset-registry/datasets/ds_8b2e7db2/versions/upload \
  -H 'X-Request-Id: req_demo_0001' \
  -F 'file=@open/demo/data/demo.csv' \
  -F 'metadata={"source":"demo"}'
```

### Create a quality rule

```bash
curl -sS -X POST http://localhost:8080/api/quality/rules \
  -H 'Content-Type: application/json' \
  -d '{"name":"basic-csv","spec":{"checks":[{"id":"cols","type":"csv_header_has_columns","columns":["id","label"],"delimiter":","}]}}'
```

### Evaluate a dataset version

```bash
curl -sS -X POST http://localhost:8080/api/quality/evaluations \
  -H 'Content-Type: application/json' \
  -d '{"dataset_version_id":"dv_3a1c0f4e"}'
```

### Create an experiment and run

```bash
curl -sS -X POST http://localhost:8080/api/experiments/experiments \
  -H 'Content-Type: application/json' \
  -d '{"name":"baseline","description":"Baseline run","metadata":{"team":"ml"}}'
```

```bash
curl -sS -X POST http://localhost:8080/api/experiments/experiments/exp_2b7f2e3b/runs \
  -H 'Content-Type: application/json' \
  -d '{"dataset_version_id":"dv_3a1c0f4e","status":"pending","params":{"lr":0.001},"metrics":{}}'
```

### Execute a training run

```bash
curl -sS -X POST http://localhost:8080/api/experiments/experiments/runs:execute \
  -H 'Content-Type: application/json' \
  -d '{"experiment_id":"exp_2b7f2e3b","dataset_version_id":"dv_3a1c0f4e","image_ref":"ghcr.io/example/train@sha256:aaaaaaaa"}'
```

### Stream live telemetry (SSE)

```bash
curl -N http://localhost:8080/api/experiments/experiment-runs/run_1f2d3a4b/stream
```

## Related docs

- `06-cli-and-usage.md`
- `03-deployment.md`
- `07-evidence-format.md`
