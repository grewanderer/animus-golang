# 14. Glossary

## Artifact

A persisted output produced by a Run, including logs, metrics, files, and model binaries. Artifact is bound to a Run and scoped to a Project.

## AuditEvent

An append-only record of a significant action or state change, used for security and compliance. AuditEvent is immutable and exportable.

## CodeRef

A reference to source code, identified by repository URL and commit SHA. CodeRef is immutable and required for production-run.

## Control Plane

The management plane responsible for metadata, policy enforcement, orchestration, and audit. Control Plane never executes user code.

## Data Plane

The execution plane that runs user code in isolated environments and provides controlled access to data and Artifact.

## Dataset

A registered logical collection of data within a Project.

## DatasetVersion

An immutable version of a Dataset with explicit metadata, schema references, and lineage. DatasetVersion is required for Run inputs.

## Developer Environment

The controlled workspace for interactive ML development, including notebooks, terminals, and remote IDE access, governed by Control Plane policies.

## EnvironmentDefinition

A reusable description of a logical execution environment, including base image, dependencies, and resource characteristics.

## EnvironmentLock

An immutable, verifiable snapshot of an execution environment, including image digests and dependency checksums.

## Model

A registered ML model within a Project, managed under explicit lifecycle and governance policies.

## ModelVersion

An immutable version of a Model linked to a source Run and Artifact.

## Pipeline

A directed acyclic graph of execution steps defined by a Pipeline Specification.

## Pipeline Specification

A declarative contract that defines pipeline steps, dependencies, resources, and error policies. It does not contain executable code.

## PipelineRun

An execution instance of a Pipeline, composed of multiple Run and governed by orchestration policy.

## Production-run

A Run executed under production governance constraints, including explicit CodeRef commit SHA and EnvironmentLock.

## Project

The primary isolation boundary and container for all domain entities in Animus Datalab.

## Run

A single execution of user code with explicit bindings to DatasetVersion, CodeRef, and EnvironmentLock.

## Service Account

A non-human principal used for automation and CI/CD, subject to Project RBAC and audit.
