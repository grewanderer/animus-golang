# Troubleshooting (Open Integration Scope)

This guide focuses on integration issues when using the SDKs and API against a running Animus gateway.

## Gateway unreachable

Symptoms:
- `connection refused` or timeouts on API requests.

Fix:
- Verify `ANIMUS_GATEWAY_URL` or `GATEWAY_URL`.
- Confirm network access from CI/training environments to the gateway.

## 401 Unauthorized

Symptoms:
- API calls return `401`.

Fix:
- Ensure a valid bearer token is supplied in the `Authorization` header.
- Confirm the token is issued by the deployment's identity provider.

## 403 Forbidden

Symptoms:
- API calls return `403`.

Fix:
- Confirm the token has the required role or scopes for the endpoint.
- Policy approval endpoints require elevated permissions.

## CI webhook signature mismatch

Symptoms:
- CI webhook endpoints return `signature_invalid`.

Fix:
- Ensure `ANIMUS_CI_WEBHOOK_SECRET` matches the gateway configuration.
- Verify the timestamp and request body used for signing.

## Run token expired or invalid

Symptoms:
- Training container calls return `401` or `unauthenticated`.

Fix:
- Ensure the job starts soon after run execution.
- Request a new run token if the job was delayed.

## Quality gate blocks downloads

Symptoms:
- Dataset download returns `quality_rule_not_set`, `quality_not_evaluated`, or `quality_gate_failed`.

Fix:
- The closed-core services enforce quality gates; request a successful evaluation for the dataset version.

## Related docs

- [05-api.md](05-api.md)
- [06-cli-and-usage.md](06-cli-and-usage.md)
- [07-evidence-format.md](07-evidence-format.md)
