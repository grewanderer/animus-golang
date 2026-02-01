# 01. System Definition and Goals

## 01.1 System definition

Animus Datalab is an enterprise digital laboratory for machine learning, intended to organize the full ML development lifecycle in a managed and reproducible form.

The platform unifies:

- data work;
- experiments;
- model training and evaluation;
- preparation of models for production use

within a single operational context with common execution, security, and audit rules.

## 01.2 Platform goals

1. Ensure reproducibility of ML experiments and results.
2. Represent the full model development context (data, code, environment, parameters, decisions) as an explicit and connected system of records.
3. Provide a managed working environment for ML developers without violating enterprise requirements.
4. Provide governance, audit, and security by default.

## 01.3 System boundaries

Animus Datalab is not:

- a source code version control system;
- an IDE as a product;
- a full inference platform.

The platform can integrate with external SCM (Git), IDEs as managed environments, and external deployment and serving systems. Detailed exclusions are consolidated in Section 13.

## 01.4 Architectural invariants

The following invariants are mandatory and must not be violated:

1. Control Plane does not execute user code.
2. Any production-run is uniquely defined by data version, code commit SHA, and a locked environment.
3. All significant actions are recorded as AuditEvent.
4. Data, code, environments, and results are explicit, versioned entities.
5. The system has no hidden state that affects execution results.
