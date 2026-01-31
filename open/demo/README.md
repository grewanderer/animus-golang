# Demo dataset

Control, not a demo. This directory contains the deterministic dataset used by the demo runner.
Requirements: docker compose (or docker-compose), curl (preferred) or python3.

Recommended quickstart:

```bash
make demo
```

Direct runner invocation (requires a running gateway):

```bash
go run ./open/cmd/demo -gateway http://localhost:8080 -dataset open/demo/data/demo.csv
```

The runner creates:
- Dataset and immutable dataset version
- Quality rule and evaluation
- Experiment and experiment run
- Lineage edges
- Audit events
