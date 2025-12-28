# Deployment and Configuration (Open Integration Scope)

This repository contains the open integration layer (schemas, SDKs, demos, and docs). The closed-core services, UI, and deployment assets are not included here.

## Closed-core deployment

Deployment of the gateway and control-plane services requires the closed-core distribution. If you need to deploy Animus DataPilot, use the private deployment guides provided with your engagement.

## SDK configuration (open)

The SDKs in this repo require only gateway access and, optionally, an auth token.

### Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `ANIMUS_GATEWAY_URL` | `http://localhost:8080` | Gateway base URL |
| `ANIMUS_AUTH_TOKEN` | empty | Bearer token for gateway auth |
| `ANIMUS_CI_WEBHOOK_SECRET` | empty | Shared secret for CI webhook signing |
| `DATAPILOT_URL` | empty | Run-scoped gateway URL provided to training containers |
| `RUN_ID` | empty | Run identifier provided to training containers |
| `TOKEN` | empty | Run-scoped bearer token provided to training containers |
| `DATASET_VERSION_ID` | empty | Dataset version identifier (training containers) |

## Demo client

Run the demo client against an existing gateway:

```bash
go run ./open/cmd/demo -gateway http://localhost:8080 -dataset open/demo/data/demo.csv
```

## Related docs

- `05-api.md`
- `06-cli-and-usage.md`
- `02-security-and-compliance.md`
