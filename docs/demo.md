# Deterministic demo scenario

The demo runner exercises the full DataPilot flow end-to-end:

1. Dataset registry: create dataset + upload immutable dataset version
2. Quality: create rule + evaluate dataset version (PASS)
3. Experiments: create experiment + create run (blocked unless quality gate passes)
4. Lineage: verify dataset/version/run/git edges exist
5. Audit: verify correlated audit events exist for the shared request ID

## Run locally

Start the stack (gateway + services + infra):

```bash
make dev
```

In a second terminal, run the demo:

```bash
go run ./cmd/demo -dataset demo/data/demo.csv
```

Optional flags / env vars:

- `-gateway` / `ANIMUS_GATEWAY_URL` (default `http://localhost:8080`)
- `-token` / `ANIMUS_BEARER_TOKEN` (required for gateway `AUTH_MODE=oidc`)
- `-request-id` / `ANIMUS_DEMO_REQUEST_ID` (audit correlation)

## Inspect results

After the runner completes it prints the `/app/...` control-plane paths to inspect:

- Dataset, version, evaluation, experiment, run
- Lineage graph rooted at the run
- Audit filtered by `request_id`

