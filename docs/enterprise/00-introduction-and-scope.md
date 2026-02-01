# 00. Introduction and Scope

## 00.1 Purpose and normative status

This document defines the target production-grade state of Animus Datalab. It specifies architectural and domain invariants, security and operational constraints, responsibilities, and failure behavior for regulated on-premise and air-gapped deployments.

This document is the normative basis for:

- architecture and security review;
- enterprise onboarding and compliance assessment;
- acceptance and audit criteria;
- operational readiness evaluation.

## 00.2 Audience and prerequisites

This document is intended for readers familiar with:

- machine learning systems;
- distributed systems;
- containerization and Kubernetes;
- enterprise security and access control.

This document is not a tutorial and does not provide usage guidance.

## 00.3 Document status and change control

This specification describes the target state of the system and is not tied to a specific implementation release.

Architectural changes are captured in Architecture Decision Records (ADR). ADRs document decisions prospectively and do not retroactively change the meaning of this specification.

All specification documents use a consistent status lifecycle: Draft -> Review -> Approved.

## 00.4 Scope and applicability

This specification applies to Animus Datalab deployments in on-premise, private cloud, and air-gapped environments.

Canonical terminology is defined in the glossary (Section 14). Non-goals and explicit exclusions are consolidated in Section 13.
