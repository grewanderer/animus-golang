# ADR 0003: Execution Contracts and Determinism

Date: 2026-01-31
Status: Accepted

## Context
Animus requires deterministic, replayable execution without embedding runtime behavior
in the Control Plane. Execution contracts must be explicit, immutable, and auditable.

## Decision
- PipelineSpec is a reusable template that declares steps, dependencies, and resources.
  It references datasets by datasetRef (logical alias) and does not bind versions.
- RunSpec is the immutable execution snapshot that binds:
  - datasetRef -> datasetVersionId
  - codeRef (repoUrl + commitSha)
  - envLock (image digests and envHash)
- Determinism is guaranteed by RunSpec immutability and strict validation of bindings.
- Validators enforce reproducibility constraints (digest-pinned images, DAG correctness,
  required bindings, and commit SHA presence).

## Consequences
- PipelineSpec can be reused across runs without changing the contract.
- RunSpec is the unit of replay and audit for execution history.
- Control Plane remains free of user code execution; runtime selection happens later.
