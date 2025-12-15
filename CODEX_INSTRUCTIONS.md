# Animus DataPilot – Codex Execution Instructions

You are implementing a production-grade enterprise pilot.
This is NOT a demo, NOT a prototype, NOT a mock.

## Absolute Rules
- No TODOs
- No stub implementations
- No fake data unless explicitly required for demo runner
- Every service must compile, run, and be testable
- Every API must be documented
- UI must render real data from backend

## Definition of Done (mandatory)
Each task is complete only if:
- make test passes
- make lint passes
- migrations apply cleanly
- API is reachable and secured
- errors are handled, logged, and audited

## Architecture Principles
- UI = control plane (single source of truth)
- SDK + CI = execution plane
- All data and metadata are immutable
- No outbound network calls required
- On-prem first, cloud optional

## Mandatory Makefile Targets
- make bootstrap
- make dev
- make test
- make lint
- make migrate-up
- make migrate-down
- make helm-lint
- make e2e

## Services to Implement
- gateway
- dataset-registry
- quality
- experiments
- lineage
- audit
- ui

## What NOT to Implement
- AutoML
- Annotation pipelines
- Kafka / Spark / Databricks
- Multi-tenant SaaS billing
- External SaaS dependencies

## Execution Order
Follow roadmap.pilot.json strictly.
Do not start next phase until:
- all tasks in current phase are done
- verification commands succeed

## Reporting
After each phase, output:
1. List of files created or changed
2. Verification commands executed
3. Confirmation of DoD compliance
