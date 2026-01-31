# Demo quickstart

Control, not a demo. The open demo uses local services and a deterministic client runner.

## Scope
- Runs the control plane locally.
- Executes the demo runner against the gateway.
- Produces deterministic dataset, quality, lineage, and audit records.

## Quickstart

```bash
make demo
```

Expected output includes:
- "==> services are running"
- "==> animus demo (gateway=...)"
- "==> smoke check ok"

## Requirements
- Go 1.22+
- Docker and docker compose (or docker-compose)
- curl (preferred) or python3

## Health checks
- Gateway: `http://localhost:8080/healthz`
- Services: `/api/*/healthz` through the gateway

## Limitations
- Demo runs in dev mode with local Postgres and MinIO.
- No external data flow; all data remains on the local machine.
- No user code execution.

## Troubleshooting
- Port conflicts: set `ANIMUS_GATEWAY_PORT`.
- Docker not running: start Docker or Docker Desktop.
- Long startup: wait for migrations to complete.
