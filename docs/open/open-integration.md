# Animus DataPilot - Open Integration Documentation

> **Open integration documentation for Animus DataPilot**  
> A deterministic, on-prem control plane for governed ML execution, lineage, and auditability.

This directory contains the **public documentation for the open integration layer** of Animus DataPilot:
APIs, SDKs, CLI usage, evidence formats, and operational guidance.

The proprietary control plane (UI, orchestration, policy engine, and backend services) is **not included** in this repository.

---

## Start here

If you are new to Animus, follow this order:

1. **Overview** -> [`00-overview.md`](00-overview.md)  
2. **Architecture** -> [`01-architecture.md`](01-architecture.md)  
3. **Security & compliance model** -> [`02-security-and-compliance.md`](02-security-and-compliance.md)

These documents explain *what the system is*, *why it exists*, and *how it is designed*.

---

## Recommended reading paths

### Evaluating Animus (architecture & trust)
For platform, security, or compliance reviewers:

- [`00-overview.md`](00-overview.md)
- [`01-architecture.md`](01-architecture.md)
- [`02-security-and-compliance.md`](02-security-and-compliance.md)

---

### Building an integration
For engineers integrating CI pipelines or tools:

- [`05-api.md`](05-api.md)
- [`06-cli-and-usage.md`](06-cli-and-usage.md)
- [`07-evidence-format.md`](07-evidence-format.md)

---

### Operating in production
For operators and platform teams:

- [`03-deployment.md`](03-deployment.md)
- [`04-operations.md`](04-operations.md)
- [`08-troubleshooting.md`](08-troubleshooting.md)

---

## Documentation map

| File | Description |
|------|-------------|
| [`00-overview.md`](00-overview.md) | Product scope, goals, personas, and non-goals |
| [`01-architecture.md`](01-architecture.md) | System components, data flow, trust boundaries |
| [`02-security-and-compliance.md`](02-security-and-compliance.md) | Auth, evidence, audit posture, guarantees |
| [`03-deployment.md`](03-deployment.md) | Deployment considerations for integrations |
| [`04-operations.md`](04-operations.md) | Runtime operations, lifecycle, monitoring |
| [`05-api.md`](05-api.md) | HTTP APIs and OpenAPI entry points |
| [`06-cli-and-usage.md`](06-cli-and-usage.md) | CLI workflows and end-to-end examples |
| [`07-evidence-format.md`](07-evidence-format.md) | Evidence bundle structure & verification |
| [`08-troubleshooting.md`](08-troubleshooting.md) | Common integration issues and fixes |
| [`09-faq.md`](09-faq.md) | Frequently asked questions |
| [`10-glossary.md`](10-glossary.md) | Canonical terminology |

---

## Repository artifacts referenced here

These documentation pages refer to the following repository components:

- **OpenAPI specifications**  
  -> [`open/api/openapi/`](../../open/api/openapi/)

- **Python SDK (CI / automation)**  
  -> [`open/sdk/python/`](../../open/sdk/python/)

- **Demo CLI and sample datasets**  
  -> [`open/cmd/demo/`](../../open/cmd/demo/)  
  -> [`open/demo/`](../../open/demo/)

---

## Scope and guarantees

This documentation covers **integration contracts only**.

Included:
- Public schemas and APIs  
- SDK usage patterns  
- CLI flows  
- Evidence formats  
- Integration constraints  

Not included:
- Control plane implementation
- UI
- Policy engine internals
- Deployment automation
- Commercial features

All proprietary components operate **on top of the documented interfaces**.

---

## Stability & compatibility

To keep integrations reliable:

- File names and URLs are stable
- Breaking changes are avoided
- Semantic meaning is preserved across versions
- New features are additive where possible

If documentation changes affect integrations, they are reflected explicitly.

---

## Feedback & coordination

For security topics, coordinated disclosure, or enterprise integration discussions,
refer to the repository's [SECURITY.md](../SECURITY.md).

---

### Tip

If you are evaluating Animus for regulated or air-gapped environments, start with:

[`02-security-and-compliance.md`](02-security-and-compliance.md)

---

(c) Animus DataPilot - Open Integration Documentation
