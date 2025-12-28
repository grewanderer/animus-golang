# CLI and Usage

This document provides end-to-end usage flows using `curl`, plus SDK guidance for training containers.

## Setup

These flows require access to a running Animus gateway (closed-core deployment). Set `GATEWAY_URL` to the gateway base URL before running the commands below.

## Authentication header

- Dev mode (`AUTH_MODE=dev`): no auth header is required.
- OIDC mode: include an `Authorization` header.

Example (OIDC token):

```bash
AUTH_HEADER='-H Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9'
```

## Flow: dataset → quality → run → evidence

```bash
export GATEWAY_URL=http://localhost:8080
AUTH_HEADER=${AUTH_HEADER:-}
```

### 1) Register a dataset

```bash
dataset_id=$(curl -sS -X POST "${GATEWAY_URL}/api/dataset-registry/datasets" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d '{"name":"fraud-dataset","description":"Fraud training set","metadata":{"owner":"ml-team"}}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["dataset_id"])')

echo "dataset_id=${dataset_id}"
```

### 2) Upload an immutable dataset version

```bash
version_id=$(curl -sS -X POST "${GATEWAY_URL}/api/dataset-registry/datasets/${dataset_id}/versions/upload" \
  ${AUTH_HEADER} \
  -F 'file=@open/demo/data/demo.csv' \
  -F 'metadata={"source":"demo"}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["version_id"])')

echo "version_id=${version_id}"
```

### 3) Define and evaluate a quality rule

```bash
rule_id=$(curl -sS -X POST "${GATEWAY_URL}/api/quality/rules" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d '{"name":"basic-csv","spec":{"checks":[{"id":"cols","type":"csv_header_has_columns","columns":["id","label"],"delimiter":","}]}}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["rule_id"])')

echo "rule_id=${rule_id}"
```

```bash
curl -sS -X POST "${GATEWAY_URL}/api/quality/evaluations" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d '{"dataset_version_id":"'"${version_id}"'","rule_id":"'"${rule_id}"'"}'
```

### 4) Create an experiment

```bash
experiment_id=$(curl -sS -X POST "${GATEWAY_URL}/api/experiments/experiments" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d '{"name":"baseline","description":"Baseline run","metadata":{"team":"ml"}}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["experiment_id"])')

echo "experiment_id=${experiment_id}"
```

### 5) Execute a training run (bind code + image)

The execute call binds the run to a dataset version, git commit, and image digest.

```bash
run_payload=$(python3 - <<'PY'
import json
payload = {
  "experiment_id": "EXP_ID",
  "dataset_version_id": "VERSION_ID",
  "image_ref": "ghcr.io/example/train@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "git_repo": "git@example.local/ml/fraud",
  "git_commit": "0123456789abcdef0123456789abcdef01234567",
  "git_ref": "refs/heads/main",
  "params": {"lr": 0.001}
}
print(json.dumps(payload))
PY
)
run_payload=${run_payload//EXP_ID/${experiment_id}}
run_payload=${run_payload//VERSION_ID/${version_id}}

run_id=$(curl -sS -X POST "${GATEWAY_URL}/api/experiments/experiments/runs:execute" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d "${run_payload}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["run_id"])')

echo "run_id=${run_id}"
```

If policy approvals are required, the response status is `202` with approval IDs.

### 6) Check run status and events

```bash
curl -sS "${GATEWAY_URL}/api/experiments/experiment-runs/${run_id}" ${AUTH_HEADER}

curl -sS "${GATEWAY_URL}/api/experiments/experiment-runs/${run_id}/events?limit=20" ${AUTH_HEADER}
```

### 7) Generate an evidence bundle

```bash
bundle_id=$(curl -sS -X POST "${GATEWAY_URL}/api/experiments/experiment-runs/${run_id}/evidence-bundles" \
  ${AUTH_HEADER} \
  -H 'Content-Type: application/json' \
  -d '{}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["bundle"]["bundle_id"])')

echo "bundle_id=${bundle_id}"
```

```bash
curl -sS -o evidence.zip \
  "${GATEWAY_URL}/api/experiments/experiment-runs/${run_id}/evidence-bundles/${bundle_id}/download" \
  ${AUTH_HEADER}
```

### 8) Export execution ledger

```bash
curl -sS "${GATEWAY_URL}/api/experiments/execution-ledger?run_id=${run_id}&limit=1" ${AUTH_HEADER}
```

## Training container usage (SDK)

Inside training or evaluation containers, use the SDK with run-scoped environment variables:

```python
import os

from animus_sdk import RunTelemetryLogger, DatasetRegistryClient, ExperimentsClient

logger = RunTelemetryLogger.from_env(timeout_seconds=2.0)
logger.log_status(status="starting")
logger.log_metric(step=1, name="loss", value=0.9)
logger.close(flush=True, timeout_seconds=5.0)

datasets = DatasetRegistryClient.from_env()
datasets.download_dataset_version(dataset_version_id=os.environ["DATASET_VERSION_ID"], dest_path="/tmp/data.zip")

exp = ExperimentsClient.from_env()
exp.upload_run_artifact(kind="model", file_path="/tmp/model.bin")
```

## Deterministic demo run

```bash
go run ./open/cmd/demo -dataset open/demo/data/demo.csv
```

## Related docs

- `05-api.md`
- `07-evidence-format.md`
- `03-deployment.md`
- `08-troubleshooting.md`
