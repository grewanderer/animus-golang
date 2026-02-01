# 05. Execution Model

## 05.1 General principles

The execution model defines how Run and PipelineRun are translated from declarative descriptions into actual execution in Data Plane under Control Plane governance.

Principles:

1. Isolation by default - each Run executes in an isolated environment and does not affect other executions.
2. Declarative intent - users describe what must be done; how it is executed is determined by the platform.
3. Controllability - execution is governed by access policies, resource limits, and security rules.
4. Observability - execution progress, results, and errors are captured and available for analysis.
5. Reproducibility - execution follows domain invariants and Section 06 requirements.

## 05.2 Run lifecycle

### 05.2.1 Run creation

Run is created in Control Plane based on a user request or automated event.

On creation, Control Plane must:

1. validate user access rights;
2. validate references to Project, DatasetVersion, CodeRef, EnvironmentLock (if required);
3. apply policies (production-run, network restrictions, quotas);
4. record an AuditEvent for creation;
5. set initial Run status to `queued`.

Run is not created if any requirement fails.

### 05.2.2 Scheduling and start

After creation, Run is handled by the Control Plane orchestrator. The orchestrator:

- selects the target Data Plane;
- reserves required resources;
- builds an execution plan, including:
    - container image (digest);
    - resources (CPU/GPU/RAM);
    - network policies;
    - secrets as references, not values;
    - access points to data and Artifact.

Before sending the execution plan to Data Plane, the platform records:

- execution parameters;
- applied policies;
- scheduler version (for reproducibility).

### 05.2.3 Execution

Run is considered started after Data Plane acknowledges the execution plan.

In Data Plane:

- an isolated execution environment is created;
- the container runs with restricted privileges;
- data is provided according to permissions;
- Artifact is written through controlled interfaces.

Control Plane does not interact directly with user code and does not participate in computation.

### 05.2.4 Run completion

Run is complete when it transitions to a terminal status:

- `succeeded`;
- `failed`;
- `canceled`.

On completion, Control Plane must:

- record final status;
- bind all created Artifact to the Run;
- record an AuditEvent for completion;
- emit integration events where configured.

## 05.3 Pipeline execution

### 05.3.1 Pipeline as DAG

Pipeline (see Section 14) is defined as a directed acyclic graph (DAG) in which:

- nodes represent execution steps;
- edges represent data or control dependencies;
- each node executes as a separate Run.

Pipeline describes execution structure, not step implementation.

### 05.3.2 PipelineRun scheduling

PipelineRun is created as a composite object including:

- a reference to Pipeline Specification;
- a set of node-runs;
- global parameters;
- retry and error handling rules.

Control Plane must:

- validate DAG correctness (no cycles);
- determine execution order;
- apply policies at PipelineRun and node levels.

### 05.3.3 Pipeline error handling

PipelineRun may complete:

- successfully if all required nodes succeed;
- with error if a critical node fails;
- partially successful if policy allows degradation.

Error handling policies must be explicitly defined in the specification.

## 05.4 Retries, reruns, and idempotency

### 05.4.1 Execution repeats

The platform supports:

- retry - automatic repeat on transient errors;
- rerun - repeat with the same inputs;
- replay - repeat execution based on a saved execution plan.

Each repeat is recorded as a separate Run with an explicit link to the original.

### 05.4.2 Control Plane idempotency

Control Plane operations must be idempotent where possible:

- Run creation with the same request-id must not produce duplicates;
- repeated status confirmations must not corrupt state;
- repeated event delivery must not break consistency.

## 05.5 Isolation and resource management

### 05.5.1 Execution isolation

Each Run executes:

- in a separate container;
- within a Project context;
- with its own rights and secrets.

The following are not permitted:

- direct access from Run to Control Plane metadata;
- shared state between Run without explicit representation.

### 05.5.2 Resource management

Each Run must define:

- CPU and RAM limits;
- GPU requirements (if applicable);
- ephemeral storage limits.

The platform must:

- prevent limit breaches;
- record resource usage;
- ensure fair scheduling across Project.

## 05.6 Errors and degradation

### 05.6.1 Error classes

Animus distinguishes the following error classes:

1. User errors - user code or configuration errors.
2. Data errors - data does not match expectations.
3. Environment errors - environment or dependency failures.
4. Platform errors - infrastructure or Control Plane failures.
5. Policy violations - violations of rules and constraints.

Each error class must be reflected explicitly in Run status and diagnostics.

### 05.6.2 Degraded operation

On partial component unavailability, the system must:

- preserve metadata consistency;
- move Run into diagnosable states (`unknown`, `reconciling`);
- allow state recovery after resolution;
- record causes of degradation in audit.

## 05.7 Execution observability

### 05.7.1 Logs

- Execution logs are collected centrally.
- Logs are bound to Run and PipelineRun steps.
- Secrets and sensitive data must not appear in logs.

### 05.7.2 Metrics

For each Run and PipelineRun, the following must be available:

- execution metrics;
- user metrics;
- resource usage metrics.

Metrics are used for analysis and decisions (for example, ModelVersion promotion).

### 05.7.3 Tracing

The platform must support tracing for:

- Control Plane operations;
- Control Plane to Data Plane interactions;
- long-running or distributed PipelineRun.

## 05.8 Execution model and audit

Each significant execution stage must produce AuditEvent, including:

- Run creation and start;
- policy enforcement;
- status transitions;
- errors and cancellations;
- execution completion.

Audit is part of the execution model, not a side effect.
