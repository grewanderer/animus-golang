# Lineage

## Overview

Lineage is captured as immutable events (`lineage_events`) with a `(subject → predicate → object)` model.

The lineage service exposes:

- event listing with filters (`GET /events`)
- lightweight BFS subgraph queries rooted at datasets, dataset versions, runs, or git commits

In most deployments, access is via the gateway:

- Service base: `http://localhost:8084`
- Gateway base: `http://localhost:8080/api/lineage`

## Predicates Used

Current core predicates emitted by services:

- `dataset has_version dataset_version`
- `experiment has_run experiment_run`
- `dataset_version used_by experiment_run`
- `experiment_run built_from git_commit`

## Subgraph Queries

Subgraph endpoints return:

- `root`: the requested node
- `nodes`: unique nodes discovered
- `edges`: lineage events connecting nodes

Query parameters:

- `depth` (default `3`, max `5`)
- `max_edges` (default `2000`, max `5000`)

