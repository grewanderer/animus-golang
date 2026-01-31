# Full-surface demo

Control, not a demo. This directory contains the deterministic dataset and userspace runner used by the full-surface demo.
Requirements: docker compose (or docker-compose), curl (preferred) or python3.

Quickstart:

```bash
make demo
```

Smoke check:

```bash
make demo-smoke
```

Components:
- `docker-compose.yml`: control plane + userspace demo stack.
- `data/`: deterministic demo dataset.
- `userspace/`: allowlisted data plane simulation container.

The demo runs:
- Project creation and scoping
- Dataset registry + dataset version upload
- Artifact presign + upload
- Run create -> plan -> dry-run -> derived state
- Audit export
- Userspace execution simulation (no user code in control plane)
