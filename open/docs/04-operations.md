# Operations (Open Integration Scope)

This repository does not include the closed-core services, UI, or deployment tooling. Operational guidance for running the control plane is distributed separately with the closed-core delivery.

## Integration considerations

- Ensure the gateway endpoint is reachable from CI and training containers.
- Manage auth tokens and CI webhook secrets in a secure secret store.
- Treat run-scoped tokens as short-lived secrets and avoid logging them.

## Evidence verification

Evidence bundles and execution ledgers are retrieved over the gateway API and verified using documented schemas.

## Related docs

- `05-api.md`
- `07-evidence-format.md`
- `08-troubleshooting.md`
