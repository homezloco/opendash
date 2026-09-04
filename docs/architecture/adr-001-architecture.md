# ADR 001: Initial Architecture

## Status

Accepted for the initial milestone; read-only details are superseded by [Architecture and current state](current-state.md).

## Context

The first milestone established boundaries between a Go API, React/Vite frontend, catalog, schemas, runtime abstraction, and deployment artifacts. OpenDash has since added SQLite, authentication, and Docker Compose lifecycle mutation.

## Decision

Retain the layered structure:

- `cmd/opendash` is the API entry point.
- `internal/api` exposes HTTP handlers and middleware.
- `internal/store` persists state in SQLite; `internal/data` provides optional demo seeds.
- `internal/runtime` abstracts Docker/Compose behind a runtime interface with a no-op fallback.
- `internal/catalog` loads the local catalog, validates manifests, plans installs, and renders Compose files.
- `web/` is a separate React SPA that proxies `/api` during development.
- `catalog/`, `schemas/`, and `api/openapi.yaml` define catalog and API contracts.
- `deploy/` contains deployment examples.

The frontend uses the real API by default and supports an explicit mock mode. The backend does not embed or serve the frontend.

## Consequences

Docker mutation is implemented in the API process and requires explicit daemon authorization; it can be disabled by configuration. SQLite and operation records are persistent, while operation execution remains in-process. Remote catalogs, updates, backup/recovery, signature verification, and isolated privileged helpers remain outside the current implementation.
