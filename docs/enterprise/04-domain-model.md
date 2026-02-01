# 04. Domain Model

## 04.1 Domain model principles

The domain model defines the set of first-class entities, their relationships, invariants, and lifecycles that formalize ML development as a system.

Principles:

1. Explicitness - all material elements are represented as entities.
2. Versioning - entities that affect results are immutable or have explicit change history.
3. Context isolation - all entities exist within a Project; implicit shared state across Project is not allowed.
4. Connectivity - entities are linked explicitly; context reconstruction does not require manual recovery.
5. Auditability - operations on entities produce AuditEvent and are recoverable after the fact.

## 04.2 Project

Project is defined in the glossary (Section 14). The following specifies attributes, relationships, invariants, and lifecycle.

### 04.2.1 Attributes

Minimum attributes:

- `project_id`;
- `name`;
- `description`;
- `owner`;
- `created_at`;
- `status` (active / archived).

### 04.2.2 Relationships

Project is the root container for:

- Dataset / DatasetVersion;
- EnvironmentDefinition / EnvironmentLock;
- Run / PipelineRun;
- Artifact;
- Model / ModelVersion;
- AuditEvent (for Project-scoped events).

### 04.2.3 Invariants

- All domain entities must belong to exactly one Project.
- No implicit dependencies may exist across Project.
- Access to entities outside a Project is prohibited.

### 04.2.4 Lifecycle

- `active` - Project is available for work.
- `archived` - Project is read-only; new Run are prohibited; history is preserved.

## 04.3 Dataset and DatasetVersion

Dataset and DatasetVersion are defined in the glossary (Section 14).

### 04.3.1 Attributes

Dataset:

- `dataset_id`;
- `name`;
- `description`;
- `owner`;
- `access_policy`.

DatasetVersion:

- `dataset_version_id`;
- `dataset_id`;
- `created_at`;
- `source_ref` (URI + source type);
- `schema_ref`;
- `statistics_ref`;
- `lineage`;
- `retention_policy`.

### 04.3.2 Invariants

- DatasetVersion is immutable.
- Any data usage in a Run must reference a specific DatasetVersion.
- Data changes are represented only by creating a new DatasetVersion.
- DatasetVersion cannot be deleted if referenced by Run, ModelVersion, or AuditEvent unless allowed by retention policy.

### 04.3.3 DatasetVersion lifecycle

- `created` -> `available`;
- `deprecated` (optional, without deletion);
- `expired` / `deleted` (only per policy and legal hold).

## 04.4 CodeRef

CodeRef is defined in the glossary (Section 14).

### 04.4.1 Attributes

- `repo_url`;
- `commit_sha`;
- `path` (optional);
- `scm_type` (for example, GitHub, GitLab, on-prem).

### 04.4.2 Invariants

- Production-run must use CodeRef with `commit_sha`.
- Branches or tags are prohibited for production-run.
- CodeRef is an immutable reference.

## 04.5 EnvironmentDefinition and EnvironmentLock

EnvironmentDefinition and EnvironmentLock are defined in the glossary (Section 14).

### 04.5.1 EnvironmentDefinition

EnvironmentDefinition describes the logical execution environment:

- base image;
- dependencies;
- resource requirements;
- additional settings (GPU, CUDA, libraries).

EnvironmentDefinition can change and be versioned.

### 04.5.2 EnvironmentLock

EnvironmentLock is a fixed immutable representation of the environment used for Run execution.

### 04.5.3 EnvironmentLock attributes

- `lock_id`;
- `environment_definition_ref`;
- `image_digest`;
- `dependency_checksums`;
- `sbom_ref`;
- `created_at`.

### 04.5.4 Invariants

- Production-run must use EnvironmentLock.
- EnvironmentLock is immutable.
- EnvironmentLock must be verifiable (digest, checksum).

## 04.6 Run and PipelineRun

Run and PipelineRun are defined in the glossary (Section 14).

### 04.6.1 Run attributes

- `run_id`;
- `project_id`;
- `dataset_versions[]`;
- `code_ref`;
- `environment_lock`;
- `parameters`;
- `status`;
- `started_at`;
- `finished_at`.

### 04.6.2 Run statuses

- `queued`;
- `running`;
- `succeeded`;
- `failed`;
- `canceled`;
- `unknown` (on loss of connectivity).

### 04.6.3 Invariants

- Run cannot modify input DatasetVersion.
- Run cannot exist without Project.
- Re-execution with the same inputs must be reproducible within the determinism model (Section 06).

## 04.7 Artifact

Artifact is defined in the glossary (Section 14).

### 04.7.1 Attributes

- `artifact_id`;
- `run_id`;
- `type`;
- `storage_ref`;
- `checksum`;
- `created_at`;
- `retention_policy`.

### 04.7.2 Invariants

- Artifact is always bound to a Run.
- Artifact cannot exist outside a Project.
- Access to Artifact is controlled by Project permissions.

## 04.8 Model and ModelVersion

Model and ModelVersion are defined in the glossary (Section 14).

### 04.8.1 ModelVersion attributes

- `model_version_id`;
- `source_run_id`;
- `artifact_ref`;
- `status`;
- `created_at`;
- `approved_by` (if applicable).

### 04.8.2 ModelVersion statuses

- `draft`;
- `validated`;
- `approved`;
- `deprecated`.

### 04.8.3 Invariants

- ModelVersion must reference a Run.
- ModelVersion promotion is recorded in AuditEvent.
- Export of approved ModelVersion may be restricted by policy.

## 04.9 AuditEvent

AuditEvent is defined in the glossary (Section 14).

### 04.9.1 Attributes

- `event_id`;
- `timestamp`;
- `actor`;
- `action`;
- `object_ref`;
- `context`;
- `result`.

### 04.9.2 Invariants

- AuditEvent is append-only.
- AuditEvent cannot be modified or deleted.
- Audit data must be exportable.

## 04.10 Domain model connectivity

System context can be reconstructed as a graph:

Project
- Dataset
  - DatasetVersion
- EnvironmentDefinition
  - EnvironmentLock
- Run / PipelineRun
  - Artifact
- Model
  - ModelVersion

AuditEvent spans the graph as the time dimension.
