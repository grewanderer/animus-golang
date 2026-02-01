# Glossary

- **Dataset**: A logical collection of data versions registered in the dataset registry.
- **Dataset version**: An immutable snapshot of dataset content with a SHA256 hash and object key.
- **Quality rule**: A declarative specification for validating a dataset version.
- **Evaluation**: The result of applying a quality rule to a dataset version (`pass`, `fail`, or `error`).
- **Experiment**: A named grouping of runs (immutable metadata container).
- **Run**: An immutable record of a specific execution with parameters, metrics, and artifacts.
- **Execution**: The act of launching a training container for a run (tracked separately from the run record).
- **Execution ledger**: A canonical JSON record tying a run to dataset hash, git commit, image digest, and policy decisions.
- **Evidence bundle**: A signed ZIP archive containing ledger, lineage, audit slice, policies, and a PDF report.
- **Audit event**: An immutable record of a write action with a SHA256 integrity hash.
- **Lineage event**: An immutable edge (subject -> predicate -> object) with integrity hash.
- **Policy**: A versioned, immutable rule set used to allow or deny execution.
- **Policy approval**: An explicit admin decision to allow or deny a policy-required run.
- **Run token**: A short-lived bearer token scoped to a run and dataset version, used by training containers.
- **Training executor**: The mechanism used to launch training jobs (`docker` or `kubernetes`).

## Related docs

- [00-overview.md](00-overview.md)
- [01-architecture.md](01-architecture.md)
- [07-evidence-format.md](07-evidence-format.md)
- [02-security-and-compliance.md](02-security-and-compliance.md)
