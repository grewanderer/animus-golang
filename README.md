![Animus Datalab](docs/assets/animus-banner.png)

Deployment: On-prem / Private cloud / Air-gapped

Animus Datalab is an enterprise ML platform with a strict separation of Control Plane and Data Plane, designed for reproducible and auditable ML development in regulated environments.

The normative technical specification is maintained under `docs/enterprise/` and is the authoritative source of system invariants, constraints, and acceptance criteria.

## Documentation

Start here:

- [`docs/README.md`](docs/README.md)

Enterprise specification (normative):

- [`docs/_generated/structure_outline.md`](docs/_generated/structure_outline.md) (specification index)
- [`docs/enterprise/00-introduction-and-scope.md`](docs/enterprise/00-introduction-and-scope.md)
- [`docs/enterprise/01-system-definition-and-goals.md`](docs/enterprise/01-system-definition-and-goals.md)
- [`docs/enterprise/02-conceptual-model.md`](docs/enterprise/02-conceptual-model.md)
- [`docs/enterprise/03-architectural-model.md`](docs/enterprise/03-architectural-model.md)
- [`docs/enterprise/03-interfaces-and-contracts.md`](docs/enterprise/03-interfaces-and-contracts.md)
- [`docs/enterprise/03-architecture-decision-records.md`](docs/enterprise/03-architecture-decision-records.md)
- [`docs/enterprise/04-domain-model.md`](docs/enterprise/04-domain-model.md)
- [`docs/enterprise/05-execution-model.md`](docs/enterprise/05-execution-model.md)
- [`docs/enterprise/06-reproducibility-and-determinism.md`](docs/enterprise/06-reproducibility-and-determinism.md)
- [`docs/enterprise/07-developer-environment.md`](docs/enterprise/07-developer-environment.md)
- [`docs/enterprise/08-security-model.md`](docs/enterprise/08-security-model.md)
- [`docs/enterprise/08-rbac-matrix.md`](docs/enterprise/08-rbac-matrix.md)
- [`docs/enterprise/09-operations-and-reliability.md`](docs/enterprise/09-operations-and-reliability.md)
- [`docs/enterprise/09-operational-runbooks.md`](docs/enterprise/09-operational-runbooks.md)
- [`docs/enterprise/10-versioning-and-compatibility.md`](docs/enterprise/10-versioning-and-compatibility.md)
- [`docs/enterprise/11-risk-and-threat-model.md`](docs/enterprise/11-risk-and-threat-model.md)
- [`docs/enterprise/12-acceptance-criteria.md`](docs/enterprise/12-acceptance-criteria.md)
- [`docs/enterprise/13-non-goals-and-exclusions.md`](docs/enterprise/13-non-goals-and-exclusions.md)
- [`docs/enterprise/14-glossary.md`](docs/enterprise/14-glossary.md)

Open integration documentation (DataPilot integration layer):

- [`docs/open/open-integration-index.md`](docs/open/open-integration-index.md)
- [`docs/open/open-integration.md`](docs/open/open-integration.md)

## Architecture summary

Control Plane:

- metadata and state management;
- orchestration and execution contracts;
- policy enforcement and audit;

Control Plane never executes user code.

Data Plane:

- pipeline execution;
- data processing and model training;
- Artifact generation.

Execution is containerized and isolated. Kubernetes is the target runtime.

See: [`docs/enterprise/03-architectural-model.md`](docs/enterprise/03-architectural-model.md)

## Demo

This repository includes an open demo that demonstrates Control Plane behavior.

> The demo is not a hosted service and does not include a production Data Plane.

### Requirements

- Go 1.22+
- Docker
- Docker Compose
- `curl` (preferred) or `python3`

### Run the demo

```bash
make demo
```

### Smoke test

```bash
make demo-smoke
```

Stop with `Ctrl+C`. For cleanup: `make demo-down`.

## Repository structure

```
.
- open/            # Schemas, SDKs, demo runner
- closed/          # Control Plane services, migrations, UI, deployment assets
- api/             # Execution and API schemas
- docs/            # Documentation index and content
- docs/enterprise/ # Normative specification
- docs/open/       # Open integration docs
- docs/_generated/ # Generated documentation artifacts
```

## Security

Animus is designed for regulated and security-sensitive environments:

- SSO (OIDC / SAML)
- Project-scoped RBAC
- secret isolation
- execution sandboxing
- full audit trail

For vulnerability reporting and coordinated disclosure, see [docs/SECURITY.md](docs/SECURITY.md).

## License

See [LICENSE](LICENSE) for details.
