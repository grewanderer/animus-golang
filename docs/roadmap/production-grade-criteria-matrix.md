# Матрица критериев production‑grade

**Версия документа:** 1.0

Ниже приведено соответствие критериев production‑grade (AC‑01…AC‑10) конкретным возможностям платформы, контрактам API, тестам и документации.

| Критерий | Реализация (фичи/сущности) | Эндпоинты (OpenAPI) | Тесты/проверки | Документы |
| --- | --- | --- | --- | --- |
| AC‑01 Полный ML‑цикл в рамках Project с DatasetVersion + CodeRef + EnvironmentLock | Project, DatasetVersion, Run, CodeRef, EnvironmentLock, DP runtime, RunDispatch/DP events | `dataset-registry.yaml`, `experiments.yaml`, `dataplane_internal.yaml` | `GOCACHE=/tmp/go-cache go test ./closed/...`, M3 API‑поток dispatch→heartbeat→terminal | `docs/enterprise/04-domain-model.md`, `docs/enterprise/05-execution-model.md`, `docs/contracts/index.md` |
| AC‑02 Воспроизводимость production‑run с явными ограничениями | RunSpec, параметры, PolicySnapshot, reproducibility bundle, EnvLock digest | `experiments.yaml` (RunSpec + bundle) | unit‑тесты валидаторов, проверка bundle | `docs/enterprise/06-reproducibility-and-determinism.md`, `docs/roadmap/execution-backlog.md` |
| AC‑03 Полный аудит и экспорт | AuditEvent append‑only, экспорт NDJSON/SIEM | `audit.yaml`, `experiments.yaml` | audit regression suite, экспорт | `docs/enterprise/08-security-model.md`, `docs/enterprise/09-operations-and-reliability.md` |
| AC‑04 SSO + RBAC на все операции | OIDC/SAML, RBAC матрица, object‑level checks | `gateway.yaml`, все сервисы | security regression suite | `docs/enterprise/08-rbac-matrix.md` |
| AC‑05 Секреты с TTL и редактированием | Secrets manager, redaction pipeline | `experiments.yaml` (execution), DP протокол | security suite (redaction) | `docs/enterprise/08-security-model.md` |
| AC‑06 Авто‑установка/апгрейд/rollback, air‑gapped | Helm/Kustomize, миграции, SBOM | Helm charts, values schema | upgrade/rollback tests | `docs/enterprise/09-operations-and-reliability.md`, `docs/enterprise/10-versioning-and-compatibility.md` |
| AC‑07 Backup/DR с RPO/RTO | backup tooling, restore drill | N/A | DR game‑day | `docs/enterprise/09-operations-and-reliability.md` |
| AC‑08 Метрики/логи/трейсы CP+DP | базовые healthz/readyz, структурные логи, фундамент для OTel/Prometheus | `dataplane_internal.yaml`, сервисные health endpoints | smoke‑проверка health, дальнейшая observability‑suite в M9 | `docs/enterprise/09-operations-and-reliability.md`, `docs/roadmap/execution-backlog.md` |
| AC‑09 DevEnv без обхода политики | DevEnv controller + audit | DP протокол + CP APIs | e2e DevEnv | `docs/enterprise/07-developer-environment.md` |
| AC‑10 Отсутствие скрытого состояния | явные сущности и связи, lineage/audit | `lineage.yaml`, `audit.yaml`, `experiments.yaml` | reproducibility bundle verification | `docs/enterprise/04-domain-model.md`, `docs/enterprise/06-reproducibility-and-determinism.md` |
