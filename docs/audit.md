# Audit

## Overview

All services emit immutable audit events into the shared `audit_events` table.

The audit service exposes read APIs for investigation and compliance:

- `GET /events` with filters and pagination
- `GET /events/{event_id}`

In most deployments, access is via the gateway:

- Service base: `http://localhost:8085`
- Gateway base: `http://localhost:8080/api/audit`

## Pagination

Use `before_event_id` as an exclusive cursor to page backwards (newest → oldest).

## Common Filters

- `actor`
- `action`
- `resource_type`
- `resource_id`
- `request_id`

