# Animus DataPilot Python SDK

This directory contains the Python SDK used by CI and pipelines to publish metadata to Animus DataPilot.

## Environment Variables

- `ANIMUS_GATEWAY_URL` (default: `http://localhost:8080`)
- `ANIMUS_AUTH_TOKEN` (optional, `Bearer` token for gateway auth)
- `ANIMUS_CI_WEBHOOK_SECRET` (optional; required if using `post_ci_webhook`)

## Experiments Usage

```python
from animus_sdk import ExperimentsClient

client = ExperimentsClient(gateway_url="http://localhost:8080")

exp = client.create_experiment(
    name="baseline",
    description="Baseline training run",
    metadata={"team": "ml", "project": "fraud"},
)

run = client.create_run(
    experiment_id=exp["experiment_id"],
    dataset_version_id="YOUR_DATASET_VERSION_ID",
    status="succeeded",
    params={"lr": 1e-3},
    metrics={"accuracy": 0.91},
)

client.post_ci_webhook(
    payload={
        "run_id": run["run_id"],
        "provider": "github_actions",
        "context": {"workflow": "train.yml", "job": "train"},
    }
)
```
