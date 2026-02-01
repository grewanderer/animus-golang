# 11. Risk and Threat Model

## 11.1 Scope

This section enumerates the assets, threat actors, and key risk categories for Animus Datalab. It defines the minimum threat model required for security and compliance review.

## 11.2 Protected assets

Protected assets include:

- data and DatasetVersion;
- Model and ModelVersion;
- Artifact and execution outputs;
- metadata and policies;
- credentials and secrets;
- audit history.

## 11.3 Threat actors

Threat actors include:

- legitimate users making errors or misconfigurations;
- malicious or compromised users;
- compromised service accounts;
- untrusted user code executed in Data Plane;
- infrastructure compromise or operator error.

## 11.4 Threat categories (STRIDE)

Minimum threat categories and required mitigations:

- Spoofing: mitigate via SSO, service account controls, and authentication verification.
- Tampering: mitigate via immutable records, integrity checks, and audit.
- Repudiation: mitigate via append-only AuditEvent and correlation identifiers.
- Information disclosure: mitigate via Project-scoped RBAC, isolation, and secret handling.
- Denial of service: mitigate via quotas, rate limits, and resource controls.
- Elevation of privilege: mitigate via least privilege and explicit policy approvals.

Security controls are defined in Section 08.

## 11.5 Operational risks and response

The following risks must have defined operational responses (see Section 09.10):

- Control Plane unavailable.
- Run stuck in queued or not starting.
- Data Plane unavailable or degraded.
- Audit export failure.
- Account compromise.
- Data or Model leakage incident.
- Control Plane metadata loss.

## 11.6 Residual risk and acceptance

Residual risk is accepted only when mitigations are documented, auditability is preserved, and acceptance criteria are met (Section 12).
