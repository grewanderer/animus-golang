# 09.10 Operational Runbooks

## 09.10.1 Scope

Operational runbooks define required response procedures for platform reliability and security incidents. Each runbook must include detection, containment, recovery, and audit steps.

## 09.10.2 Required runbooks

RB-01: Control Plane unavailable.

RB-02: Run does not start or remains in `queued`.

RB-03: Data Plane unavailable or degraded.

RB-04: Audit export failure or delivery backlog.

RB-05: Suspected account compromise.

RB-06: Data or Model leakage incident.

RB-07: Control Plane metadata loss or corruption.

## 09.10.3 Cross-references

- Security escalation and audit requirements: [08-security-model.md](08-security-model.md)
- Recovery and backup requirements: [09-operations-and-reliability.md](09-operations-and-reliability.md)
- Acceptance criteria for operational readiness: [12-acceptance-criteria.md](12-acceptance-criteria.md)
