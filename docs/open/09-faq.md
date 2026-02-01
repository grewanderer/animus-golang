# FAQ

These questions describe the Animus DataPilot control plane (closed core) for integration context.

## Does Animus DataPilot run air-gapped?

Yes. The system is designed for on-prem and air-gapped deployments. The only required network paths are internal (gateway <-> services, services <-> Postgres/MinIO). OIDC can be configured with an internal IdP.

## Does Animus run training logic?

No. Training and evaluation logic are fully contained in user-owned containers. Animus only launches containers and records metadata.

## Can I use external object storage instead of MinIO?

Yes. Any S3-compatible endpoint can be used via `ANIMUS_MINIO_*` configuration.

## How are datasets made immutable?

Dataset versions are stored as immutable objects in MinIO with content SHA256 hashes. Version metadata is persisted in Postgres and never mutated.

## How do quality gates work?

A dataset version is blocked from download and training unless the latest evaluation status is `pass`.

## Is there a built-in feature store or AutoML?

No. AutoML, feature stores, annotation pipelines, and SaaS integrations are out of scope for this repository.

## Are rate limits or quotas built in?

No. Enforce rate limits and quotas at ingress or via Kubernetes `ResourceQuota`.

## How do I prove what code and image were used for a run?

Use the execution ledger and evidence bundle. The ledger records git commit and image digest, and the evidence bundle includes a signed copy.

## Is mTLS supported between services?

Not implemented in this repository. Use a service mesh or ingress with mTLS if required.

## Related docs

- [02-security-and-compliance.md](02-security-and-compliance.md)
- [07-evidence-format.md](07-evidence-format.md)
- [03-deployment.md](03-deployment.md)
- [10-glossary.md](10-glossary.md)
