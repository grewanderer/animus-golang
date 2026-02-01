# 03.7 Architecture Decision Records (ADR)

ADR capture key architectural decisions that affect system invariants, security, reproducibility, and operations. ADR are part of the architectural contract.

## ADR-001: Control Plane and Data Plane separation

**Status:** Accepted

**Date:** 2026-02-01

**Impacted areas:** Architecture, Security, Operations

### Context

Animus executes user ML code in enterprise environments with security, audit, and reproducibility requirements. User code cannot be treated as trusted.

### Decision

The system is split into two logically and infrastructurally independent planes: Control Plane and Data Plane (see Section 14). Control Plane handles management, metadata, orchestration, and audit; Data Plane executes user code and handles data.

Control Plane never executes user code.

### Alternatives considered

1. Monolithic system with shared execution.
2. Partial execution in Control Plane.
3. In-process execution (rejected).

### Rationale

- eliminates RCE-class attacks in the management plane;
- simplifies security review;
- enables independent scaling;
- provides a predictable failure model.

### Consequences

- increased architectural complexity;
- need for strict contracts between planes;
- higher observability requirements.

### Related sections

- Section 03 (Architectural Model)
- Section 08 (Security Model)
- Section 09 (Operations and Reliability)

---

## ADR-002: Run as the unit of reproducibility

**Status:** Accepted

**Impacted areas:** Architecture, DX, Compliance

### Context

ML experiments are traditionally non-reproducible due to implicit execution context.

### Decision

Run is introduced as the minimal unit of meaning, reproducibility, and audit.

Run is defined exclusively by:

- DatasetVersion;
- CodeRef (`commit_sha`);
- EnvironmentLock;
- parameters and policies.

### Alternatives considered

1. Experiment as a logical group.
2. Notebook-centric model.
3. Best-effort tracking.

### Rationale

- Run is formal and reproducible;
- Run applies to batch and pipeline execution;
- Run is a convenient audit and analysis anchor.

### Consequences

- interactive work is not a result;
- explicit transition from dev to Run is required.

### Related sections

- Section 02 (Conceptual Model)
- Section 04 (Domain Model)
- Section 05 (Execution Model)
- Section 06 (Reproducibility and Determinism)

---

## ADR-003: Immutable DatasetVersion

**Status:** Accepted

**Impacted areas:** Data, Reproducibility, Compliance

### Context

Mutable datasets make reproducibility impossible.

### Decision

Dataset and DatasetVersion are separated (see Section 14). All Run must reference DatasetVersion.

### Alternatives considered

1. Mutable datasets with timestamps.
2. Snapshot-on-run.
3. Best-effort versioning.

### Rationale

- formal reproducibility;
- correct lineage;
- clear retention and legal hold model.

### Consequences

- increased number of versions;
- need for lifecycle policies.

### Related sections

- Section 04 (Domain Model)
- Section 06 (Reproducibility and Determinism)
- Section 08 (Security Model)

---

## ADR-004: Production-run requires commit SHA and EnvironmentLock

**Status:** Accepted

**Impacted areas:** Security, Reproducibility, Operations

### Context

Branches, tags, and floating dependencies are unacceptable for production results.

### Decision

Production-run is permitted only when:

- CodeRef includes `commit_sha`;
- EnvironmentLock is used.

### Alternatives considered

1. Allow branches with warning.
2. Auto-snapshot environments.
3. Soft enforcement.

### Rationale

- eliminates the class of "unknown what was executed";
- simplifies audit and incident analysis;
- aligns with regulated environment requirements.

### Consequences

- higher DX barrier for new users;
- need for tooling to generate locks.

### Related sections

- Section 06 (Reproducibility and Determinism)
- Section 04 (EnvironmentLock)
- Section 12 (Acceptance Criteria)

---

## ADR-005: IDE as a tool, not a platform

**Status:** Accepted

**Impacted areas:** DX, Architecture

### Context

ML developers expect IDEs, but IDEs do not solve reproducibility, security, or governance.

### Decision

Animus does not implement its own IDE. It provides managed IDE sessions as part of Developer Environment.

IDE:

- is not a source of truth;
- does not bypass policies;
- does not define the execution model.

### Alternatives considered

1. Built-in web IDE.
2. Notebook-first platform.
3. IDE as primary abstraction.

### Rationale

- reduced scope;
- leverage mature tools;
- preserve architectural clarity.

### Consequences

- requires integration with external IDEs;
- DX depends on remote tooling quality.

### Related sections

- Section 01 (System boundaries)
- Section 07 (Developer Environment)

---

## ADR-006: Audit is append-only and cannot be disabled

**Status:** Accepted

**Impacted areas:** Security, Compliance

### Context

Audit is mandatory in regulated environments.

### Decision

AuditEvent:

- is append-only;
- cannot be disabled;
- is exportable;
- is part of the execution model.

### Alternatives considered

1. Configurable audit.
2. Partial audit.
3. External-only audit.

### Rationale

- aligns with regulatory requirements;
- simplifies incident response;
- reduces gray areas.

### Consequences

- increased data volume;
- retention and storage requirements.

### Related sections

- Section 05 (Execution Model)
- Section 08 (Security Model)
- Section 09 (Operations and Reliability)

---

# 03.7.1 ADR process

## 03.7.1.1 When a new ADR is required

ADR is required if a decision:

- changes the domain model;
- affects reproducibility;
- impacts security;
- complicates rollback;
- introduces a new trust boundary.

## 03.7.1.2 When an ADR is not required

- library selection;
- internal optimization;
- UI details;
- implementation details without architectural impact.
