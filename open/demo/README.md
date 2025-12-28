# Demo scenario (deterministic)

This directory contains the minimal demo dataset used by the DataPilot demo runner (`open/cmd/demo`).

- `data/demo.csv`: small CSV with a stable header for quality checks

Run the demo runner against a running Animus gateway:

```bash
go run ./open/cmd/demo -gateway http://localhost:8080 -dataset open/demo/data/demo.csv
```

The runner creates:

- Dataset + immutable dataset version
- Quality rule + evaluation (PASS)
- Experiment + experiment run (gated by quality)
- Lineage edges (dataset → version → run → git commit)
- Audit events (correlated by a shared `X-Request-Id`)
