# ADR 0001: Control Plane and Data Plane Separation

Date: 2026-01-30
Status: Accepted

## Context
The production-grade specification requires a strict separation between the Control Plane
(API, metadata, orchestration, policy, audit) and the Data Plane (user code execution).
The Control Plane must never execute user code directly.

## Decision
- Runtime execution code is isolated under `closed/internal/runtimeexec`.
- Control Plane services must not import runtime execution packages.
- Execution is delegated to external runtimes (Kubernetes Jobs or Docker) and only
  observed via status/telemetry APIs.
- A lint guard (depguard) prevents Control Plane packages from importing
  `closed/internal/runtimeexec`.

## Consequences
- New execution backends must live in `closed/internal/runtimeexec`.
- Control Plane code should depend on interfaces and avoid direct runtime operations.
- Any attempt to call runtime execution logic from Control Plane packages will be
  blocked by lint rules.
