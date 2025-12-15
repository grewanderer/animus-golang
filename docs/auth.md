# Identity, RBAC, and Gateway Routing

## Roles

Animus uses three built-in roles:

- `viewer`: read-only access
- `editor`: write access
- `admin`: administrative access (superset of editor)

Default RBAC mapping (enforced at the gateway and service boundaries):

- `GET`, `HEAD`, `OPTIONS` → requires `viewer`
- all other methods → requires `editor`

## Auth Modes (Gateway)

Gateway authentication is controlled by `AUTH_MODE`:

- `dev`: always authenticates as a fixed identity (`DEV_AUTH_*`)
- `oidc`: validates OIDC ID tokens (Bearer header or session cookie)
- `disabled`: no authentication (intended only for development/debugging)

## Service Boundary Enforcement

Backend services are protected by gateway-signed identity headers.

Gateway injects:

- `X-Animus-Subject`
- `X-Animus-Email`
- `X-Animus-Roles`
- `X-Animus-Auth-Ts`
- `X-Animus-Auth-Sig`

Services require a shared HMAC secret (`ANIMUS_INTERNAL_AUTH_SECRET`) to verify
`X-Animus-Auth-Sig` and reject direct (non-gateway) calls.

## Required Environment Variables

Gateway + services:

- `ANIMUS_INTERNAL_AUTH_SECRET`: shared secret for gateway↔service request signing

Gateway auth (OIDC mode):

- `AUTH_MODE=oidc`
- `OIDC_ISSUER_URL`
- `OIDC_CLIENT_ID`
- Optional login flow (enables `/auth/login` + `/auth/callback`):
  - `OIDC_CLIENT_SECRET`
  - `OIDC_REDIRECT_URL`

Local development defaults:

- `make dev` sets `AUTH_MODE=dev` and generates `ANIMUS_INTERNAL_AUTH_SECRET` for the session.

