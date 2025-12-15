# Experiments & CI Integrations

## Overview

The experiments service provides:

- An immutable experiment registry (`experiments`)
- Immutable run records (`experiment_runs`)
- CI/build context attachment via a signed webhook (`experiment_run_contexts`)

In most deployments, clients call these APIs through the gateway:

- Service base: `http://localhost:8083`
- Gateway base: `http://localhost:8080/api/experiments`

## Endpoints

- `GET /experiments` (optional `?limit=...&name=...`)
- `POST /experiments`
- `GET /experiments/{experiment_id}`
- `GET /experiments/{experiment_id}/runs` (optional `?limit=...`)
- `POST /experiments/{experiment_id}/runs`
- `GET /experiment-runs/{run_id}`
- `POST /ci/webhook` (signed)

## Run Status Values

Runs use a fixed status enum:

- `pending`
- `running`
- `succeeded`
- `failed`
- `canceled`

## Quality Gate Enforcement

If an experiment run is created with `dataset_version_id`, the service enforces quality gating:

- the referenced dataset version must have `quality_rule_id` set
- the latest evaluation for `(dataset_version_id, quality_rule_id)` must have `status=pass`

If the gate is not satisfied, run creation returns `409` with:

- `quality_rule_not_set`
- `quality_not_evaluated`
- `quality_gate_failed`

## CI Webhook (Signed)

Endpoint: `POST /ci/webhook` (via gateway: `POST /api/experiments/ci/webhook`)

Required headers:

- `X-Animus-CI-Ts`: unix timestamp (seconds)
- `X-Animus-CI-Sig`: base64url(HMAC-SHA256) signature (no padding)

Signature scheme:

- `body_sha256_hex = sha256(body_bytes)`
- `message = ts + "\n" + upper(method) + "\n" + body_sha256_hex`
- `signature = base64url(hmac_sha256(ANIMUS_CI_WEBHOOK_SECRET, message))`

Required environment variable (experiments service):

- `ANIMUS_CI_WEBHOOK_SECRET`

Payload requirement:

- JSON body must include `run_id` (string). Additional fields are stored immutably.

