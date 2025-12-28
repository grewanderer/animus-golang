# Security Policy

## Reporting a vulnerability

This repository contains open integration artifacts (OpenAPI specs, SDKs, demo clients, and documentation). It does not include the closed-core control plane services or UI.

For security issues in the closed core, use the private security reporting channel established for your deployment or engagement. There is no public disclosure mailbox defined in this repository.

For security issues in the SDK or demo clients in this repo, report issues to the project maintainers through a confidential ticketing system in your organization.

If you do not have a private channel, report issues to the project maintainers through a confidential ticketing system in your organization.

## Supported versions

The supported version is the specific commit or release artifact deployed in your environment. Pin a commit for production deployments and backport fixes as required by your change control process.

## Threat model summary

The Animus DataPilot core security model (gateway-only access, RBAC, run tokens, immutable audit/lineage) is implemented in the closed-core services and is documented in the deployment-specific materials.

For integration guidance and evidence formats, see `open/docs/02-security-and-compliance.md`.
