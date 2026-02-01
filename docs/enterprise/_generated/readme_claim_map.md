# README Claim Map (pre-rewrite)

This map lists platform-property claims found in the previous README.md and their support in `docs/enterprise/**`.

| Claim (verbatim) | Status | docs/enterprise reference(s) / notes |
| --- | --- | --- |
| Deployment: On-prem / Private cloud / Air-gapped | SUPPORTED | `docs/enterprise/00-introduction-and-scope.md` §00.4 Scope and applicability; `docs/enterprise/09-operations-and-reliability.md` §09.2 Deployment model |
| Animus Datalab is an enterprise ML platform with a strict separation of Control Plane and Data Plane, designed for reproducible and auditable ML development in regulated environments. | SUPPORTED | `docs/enterprise/01-system-definition-and-goals.md` §01.1 System definition; `docs/enterprise/03-architectural-model.md` §03.1 Architecture overview; `docs/enterprise/02-conceptual-model.md` §02.4 Reproducibility as a platform property; `docs/enterprise/08-security-model.md` §08.5 Audit; `docs/enterprise/00-introduction-and-scope.md` §00.1 Purpose and normative status |
| The normative technical specification is maintained under `docs/enterprise/` and is the authoritative source of system invariants, constraints, and acceptance criteria. | AMBIGUOUS | Normative status and scope are defined in `docs/enterprise/00-introduction-and-scope.md` §00.1, but the repository path `docs/enterprise/` is not specified in the enterprise docs. |
| Control Plane: metadata and state management; | AMBIGUOUS | `docs/enterprise/03-architectural-model.md` §03.2 Control Plane (metadata management is explicit; “state management” is not stated verbatim). |
| Control Plane: orchestration and execution contracts; | SUPPORTED | `docs/enterprise/03-architectural-model.md` §03.2 Control Plane; `docs/enterprise/03-interfaces-and-contracts.md` §03.6 Interfaces and Contracts |
| Control Plane: policy enforcement and audit; | SUPPORTED | `docs/enterprise/03-architectural-model.md` §03.2 Control Plane; `docs/enterprise/08-security-model.md` §08.5 Audit |
| Control Plane never executes user code. | SUPPORTED | `docs/enterprise/01-system-definition-and-goals.md` §01.4 Architectural invariants; `docs/enterprise/03-architectural-model.md` §03.2 Control Plane; `docs/enterprise/14-glossary.md` Control Plane |
| Data Plane: pipeline execution; | SUPPORTED | `docs/enterprise/05-execution-model.md` §05.3 Pipeline execution; `docs/enterprise/03-architectural-model.md` §03.3 Data Plane |
| Data Plane: data processing and model training; | AMBIGUOUS | System-level goals include data work and model training (`docs/enterprise/01-system-definition-and-goals.md` §01.1), but Data Plane responsibilities do not explicitly name “data processing” or “model training.” |
| Data Plane: Artifact generation. | SUPPORTED | `docs/enterprise/02-conceptual-model.md` §02.2 Run produces Artifact; `docs/enterprise/03-architectural-model.md` §03.3 Data Plane; `docs/enterprise/04-domain-model.md` §04.7 Artifact |
| Execution is containerized and isolated. | SUPPORTED | `docs/enterprise/03-architectural-model.md` §03.3 Data Plane; `docs/enterprise/05-execution-model.md` §05.5 Isolation and resource management; `docs/enterprise/08-security-model.md` §08.6 Execution isolation |
| Kubernetes is the target runtime. | SUPPORTED | `docs/enterprise/03-architectural-model.md` §03.3 Data Plane (Kubernetes required execution environment); `docs/enterprise/09-operations-and-reliability.md` §09.2 Deployment model |
| This repository includes an open demo that demonstrates Control Plane behavior. | UNSUPPORTED | No corresponding statement in `docs/enterprise/**`. |
| The demo is not a hosted service and does not include a production Data Plane. | UNSUPPORTED | No corresponding statement in `docs/enterprise/**`. |
| Animus is designed for regulated and security-sensitive environments: | AMBIGUOUS | Regulated environments are in scope (`docs/enterprise/00-introduction-and-scope.md` §00.1), but “security-sensitive” is not stated verbatim. |
| SSO (OIDC / SAML) | SUPPORTED | `docs/enterprise/08-security-model.md` §08.2 Authentication |
| Project-scoped RBAC | SUPPORTED | `docs/enterprise/08-security-model.md` §08.3 Authorization and RBAC; `docs/enterprise/08-rbac-matrix.md` §08.2.1 Scope |
| secret isolation | AMBIGUOUS | Secrets handling is defined (`docs/enterprise/08-security-model.md` §08.4), but “secret isolation” is not stated verbatim. |
| execution sandboxing | AMBIGUOUS | Execution isolation is defined (`docs/enterprise/08-security-model.md` §08.6), but “sandboxing” is not stated verbatim. |
| full audit trail | SUPPORTED | `docs/enterprise/08-security-model.md` §08.5 Audit; `docs/enterprise/05-execution-model.md` §05.8 Execution model and audit; `docs/enterprise/03-architecture-decision-records.md` ADR-006 |
