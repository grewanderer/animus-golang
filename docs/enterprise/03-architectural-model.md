# 03. Architectural Model

## 03.1 Architecture overview

Animus is a distributed system with a strict separation of responsibilities between Control Plane and Data Plane.

- Control Plane manages governance, policies, metadata, audit, and orchestration.
- Data Plane executes user code in isolated environments and handles data and Artifact access.

The separation is an architectural invariant and is required for:

1. security, by preventing untrusted code execution in the management plane;
2. scaling, by independently scaling management and execution;
3. reliability, by preserving metadata and audit consistency when Data Plane fails.

## 03.2 Control Plane

Control Plane implements the management plane and provides:

- external interfaces and API contract (see Section 03.6);
- storage and management of domain metadata;
- orchestration and execution planning;
- enforcement of access, security, resource, and retention policies;
- generation of AuditEvent.

Constraints:

- Control Plane does not execute user code.
- Control Plane does not require data access sufficient for execution; references and metadata are sufficient.

## 03.3 Data Plane

Data Plane implements the execution plane and provides:

1. execution of user code in containerized environments;
2. isolation of compute and resources;
3. controlled access to data and Artifact;
4. collection of logs, metrics, and execution traces.

Baseline execution requirements:

- Kubernetes is the required execution environment;
- each Run executes in an isolated environment;
- Run resources are explicitly defined (CPU/GPU/RAM/ephemeral storage);
- Run environment is defined by EnvironmentLock;
- network policies are enforced at execution.

Data and Artifact access is provided through controlled mechanisms:

- reading DatasetVersion from allowed sources;
- writing Artifact to object storage;
- support for read-only sources;
- binding results to Project and Run.

Data Plane does not provide unaudited channels for data access outside platform policy.

Secrets are provided to execution:

- temporarily;
- in the minimal required scope;
- without exposing values in UI or logs;
- with access attempts recorded in audit.

## 03.4 Trust boundaries

Animus distinguishes the following trust zones:

1. User clients (UI/CLI/SDK) are untrusted and require strict authentication and authorization.
2. Control Plane is a trusted management zone that does not execute user code.
3. Data Plane is a partially trusted execution zone that runs untrusted code and isolates it.
4. External systems (SCM, registry, vault, storage, SIEM) are separate trust zones integrated through contractual interfaces.

Trust boundaries define requirements for network policies, least privilege, action logging, and isolation.

Security requirements and the formal threat model are specified in Sections 08 and 11.

## 03.5 Failure model (principled failure model)

Animus is designed with the assumption that failure is a normal mode of distributed systems.

Failure behavior requirements:

- Control Plane operations are idempotent where possible;
- Run status transitions must converge to consistent states (for example, `unknown` or `reconciling`) on loss of Data Plane connectivity;
- the system must provide reconciliation mechanisms that restore observed state after temporary component loss;
- failure scenarios must be observable through metrics, logs, and tracing.

Degradation principle:

- Data Plane failure must not cause loss of metadata for already created entities or audit history;
- active executions may be interrupted and marked accordingly, with preserved context and cause.
