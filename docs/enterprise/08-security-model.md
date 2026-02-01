# 08. Security Model

## 08.1 Security principles

Security is defined by the following principles:

- zero trust between components and actors;
- least privilege for all access;
- explicit policy enforcement;
- auditability of all significant actions.

## 08.2 Authentication

Authentication requirements:

- SSO via OIDC and/or SAML.
- Session time-to-live (TTL) enforced.
- Forced logout supported.
- Limits on parallel sessions enforced.

Service accounts are supported for automation and CI/CD. Service account usage is audited and subject to Project RBAC.

## 08.3 Authorization and RBAC

Authorization is Project-scoped and enforced for all domain entities. Default deny applies if no explicit permission is granted.

Object-level constraints apply to Dataset, Run, Model, and Artifact access.

The RBAC matrix and minimum role semantics are defined in [08-rbac-matrix.md](08-rbac-matrix.md).

## 08.4 Secrets management

Secrets management requirements:

- Integration with an external secret store (vault-like).
- Secrets are provided temporarily and with minimal scope.
- Secrets must not appear in UI, logs, metrics, or Artifact.
- Secret access attempts are recorded in AuditEvent.

## 08.5 Audit

Audit requirements:

- All significant actions must generate AuditEvent.
- AuditEvent is append-only and cannot be disabled.
- Audit export must support SIEM and monitoring integrations.

Audit is a security control and is required for acceptance (Section 12).

## 08.6 Execution isolation

Data Plane executes untrusted user code in containerized environments with restricted privileges.

Control Plane never executes user code. Data Plane access to Control Plane metadata is limited to the minimum required interfaces.

Network access and resource limits are enforced by policy (see Sections 03.3 and 05).
