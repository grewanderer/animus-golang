# 12. Acceptance Criteria

Acceptance criteria define the formal conditions under which Animus Datalab is considered production-grade and suitable for regulated environments. The platform is not accepted if any mandatory criterion is unmet.

## 12.1 Mandatory criteria

AC-01. Full ML lifecycle is executable within one Project and includes DatasetVersion, CodeRef commit SHA, and EnvironmentLock bindings.

AC-02. Production-run reproducibility is explicit, and any limitations are recorded in Run metadata and AuditEvent (Section 06).

AC-03. All significant actions generate AuditEvent; audit export is reliable and consistent under retries (Sections 05.8 and 08.5).

AC-04. SSO authentication and Project-scoped RBAC are enforced end-to-end, including object-level access for Dataset, Run, Model, and Artifact (Section 08).

AC-05. Secrets are temporary, never exposed in UI/logs/Artifact, and access is audited (Section 08.4).

AC-06. Installation and upgrades are automated, support air-gapped environments, and provide rollback without data loss (Sections 09 and 10).

AC-07. Backup and disaster recovery procedures are documented, testable, and effective with defined RPO/RTO (Section 09.5).

AC-08. Metrics, structured logs, and tracing are available for Control Plane, Data Plane, and Run execution paths (Sections 05.7 and 09.4).

AC-09. Developer Environment is available and does not bypass governance or policy constraints (Section 07).

AC-10. No hidden state affects results; all outcomes are explainable through explicit entities and recorded bindings (Sections 04 and 06).

## 12.2 Production-grade definition

Animus Datalab is production-grade when all mandatory criteria are satisfied and verified on a working installation with security and audit policies enabled.
