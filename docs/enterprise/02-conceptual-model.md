# 02. Conceptual Model

## 02.1 Digital laboratory for ML development

In the context of Animus, a digital laboratory is a managed computational and organizational environment in which ML development is conducted as a reproducible process with formal objects and verifiable properties.

The laboratory establishes a single operating context in which:

1. developer actions are interpreted as operations on domain objects;
2. all material dependencies are captured explicitly;
3. security and audit requirements are enforced by the platform.

## 02.2 Run as a unit of meaning

Run (see Section 14) is used as the minimal unit of execution and reproducibility.

Run is uniquely defined by the following inputs:

- Project;
- DatasetVersion (one or more);
- CodeRef with `commit_sha`;
- EnvironmentLock;
- execution parameters;
- execution policy (for example, network egress restrictions).

Run produces:

- Artifact (logs, metrics, files, reports, model weights);
- execution trace;
- AuditEvent for all significant operations and status changes.

Run is a contract: when inputs match and policy permits execution, the result must be reproducible within the determinism model (Section 06).

## 02.3 Explicit ML experiment context

Animus introduces a mandatory requirement:

All elements that affect a result must be represented in the system as explicit entities or references to entities.

Minimum explicit context:

1. DatasetVersion, schema, statistics, and lineage of data;
2. CodeRef with `commit_sha` and repository reference;
3. EnvironmentDefinition and EnvironmentLock;
4. Run and PipelineRun with specification and status history;
5. Artifact and metrics as first-class objects;
6. policies, access, and audit.

Any action that cannot be bound to explicit context is treated as a process defect and must be eliminated or made explicit.

## 02.4 Reproducibility as a platform property

Reproducibility is treated as a platform property derived from the domain model and execution constraints, not as a team discipline. Formal definitions and requirements are specified in Section 06.
