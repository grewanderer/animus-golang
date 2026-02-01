# 08.2 RBAC Matrix

## 08.2.1 Scope

RBAC is Project-scoped. Roles apply to all domain entities within a Project, with additional object-level constraints for Dataset, Run, Model, and Artifact.

Default deny applies when no explicit permission is present.

## 08.2.2 Minimum role semantics

| Role | Read APIs | Write APIs | Policy approvals |
| --- | --- | --- | --- |
| viewer | Y | N | N |
| editor | Y | Y | N |
| admin | Y | Y | Y |

Notes:

- Read operations include retrieval and list operations.
- Write operations include create, update, delete, and execute actions.
- Policy approval is required where governance rules mandate explicit approval.

## 08.2.3 Service accounts

Service accounts are non-human principals used for automation and CI/CD. They must be assigned explicit Project roles and are subject to the same RBAC constraints as users.
