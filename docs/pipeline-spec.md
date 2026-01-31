# PipelineSpec (Execution Contract)

This document defines the PipelineSpec used by Animus for deterministic, replayable execution planning.
The spec is a reusable, declarative template and is intentionally strict to prevent hidden defaults.

Schema: `api/pipeline_spec.yaml`

## Separation of Responsibilities

PipelineSpec:
- Declarative execution template.
- References datasets by datasetRef (logical alias).
- Reusable across runs.

RunSpec:
- Binds datasetRef to datasetVersionId.
- Binds commit SHA.
- Snapshots environment (image digests and env hash).
- Is the unit of determinism and replay.

## Determinism Rules

- Image references must be pinned by digest (no tags).
- All fields are explicit; empty lists are allowed but must be present.
- Environment variables are fully declared per step (no inheritance).
- Dependencies are explicit DAG edges; there is no implied ordering.
- Retry policy and resource requests are required per step.
- Determinism is achieved via RunSpec bindings, not PipelineSpec defaults.

## Top-level Fields

- `apiVersion` (string, required)
- `kind` (string, required, must be `Pipeline`)
- `specVersion` (string, required)
- `metadata` (object, optional)
  - `name` (string, required if metadata is provided)
  - `description` (string, optional)
  - `labels` (map[string]string, optional)
- `spec` (object, required)
  - `steps` (array, required)
  - `dependencies` (array, required)

## Step Definition

Each step is a fully declared execution unit:

- `name` (string, required, stable identifier)
- `image` (string, required, digest-pinned)
- `command` (array[string], required)
- `args` (array[string], required)
- `inputs` (object, required)
  - `datasets` (array, required)
    - `name` (string)
    - `datasetRef` (string, logical dataset alias; bound in RunSpec)
  - `artifacts` (array, required)
    - `name` (string)
    - `fromStep` (string)
    - `artifact` (string)
- `outputs` (object, required)
  - `artifacts` (array, required)
    - `name` (string)
    - `type` (string)
    - `mediaType` (string, optional)
    - `description` (string, optional)
- `env` (array, required)
  - `name` (string)
  - `value` (string)
- `resources` (object, required)
  - `cpu` (string, required, explicit units)
  - `memory` (string, required, explicit units)
  - `gpu` (int, required, can be 0)
- `retryPolicy` (object, required)
  - `maxAttempts` (int, required, >=1)
  - `backoff` (object, required)
    - `type` (`fixed` or `exponential`)
    - `initialSeconds` (int, >=0)
    - `maxSeconds` (int, >=0)
    - `multiplier` (number, >=1)

## Dependencies

Dependencies define the DAG edges explicitly:

- `dependencies[]`
  - `from` (string, step name)
  - `to` (string, step name)

## Example

```yaml
apiVersion: animus/v1alpha1
kind: Pipeline
specVersion: "1.0"
metadata:
  name: fraud-training
  description: Deterministic training pipeline
spec:
  steps:
    - name: ingest
      image: ghcr.io/acme/ingest@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
      command: ["python", "-m", "ingest"]
      args: ["--source", "s3://raw-bucket/data.csv"]
      inputs:
        datasets: []
        artifacts: []
      outputs:
        artifacts:
          - name: raw-data
            type: dataset
      env: []
      resources:
        cpu: "2"
        memory: "4Gi"
        gpu: 0
      retryPolicy:
        maxAttempts: 3
        backoff:
          type: exponential
          initialSeconds: 10
          maxSeconds: 300
          multiplier: 2
    - name: train
      image: ghcr.io/acme/train@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
      command: ["python", "-m", "train"]
      args: ["--epochs", "10"]
      inputs:
        datasets:
          - name: labels
            datasetRef: labels
        artifacts:
          - name: raw-data
            fromStep: ingest
            artifact: raw-data
      outputs:
        artifacts:
          - name: model
            type: model
      env:
        - name: SEED
          value: "1337"
      resources:
        cpu: "4"
        memory: "16Gi"
        gpu: 1
      retryPolicy:
        maxAttempts: 1
        backoff:
          type: fixed
          initialSeconds: 0
          maxSeconds: 0
          multiplier: 1
  dependencies:
    - from: ingest
      to: train
```
