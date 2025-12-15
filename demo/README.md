# Demo scenario (deterministic)

This directory contains the minimal demo dataset used by the DataPilot demo runner (`cmd/demo`).

- `data/demo.csv`: small CSV with a stable header for quality checks

Run the demo runner after starting the stack via `make dev`:

```bash
go run ./cmd/demo -dataset demo/data/demo.csv
```

The runner creates:

- Dataset + immutable dataset version
- Quality rule + evaluation (PASS)
- Experiment + experiment run (gated by quality)
- Lineage edges (dataset → version → run → git commit)
- Audit events (correlated by a shared `X-Request-Id`)

