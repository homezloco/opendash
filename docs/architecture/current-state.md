# Architecture and current state

## Components

- `cmd/opendash` loads environment configuration, opens SQLite, recovers interrupted operation records, selects Docker or no-op runtime, and serves the API.
- `internal/api` provides `/api/v1`, authentication/CSRF middleware, catalog/install planning, lifecycle handlers, observation, and in-process asynchronous operation workers.
- `internal/auth` stores one Argon2id administrator and hashed opaque sessions in SQLite.
- `internal/store` owns SQLite migrations and persistent dashboard, app, revision, and operation records. `internal/data` supplies optional demo seed rows.
- `internal/catalog` loads local JSON, validates runtime constraints, evaluates selected risks/conflicts, allocates localhost ports, and renders Compose/environment files.
- `internal/runtime/docker` invokes Docker/Compose for readiness, install, lifecycle, logs, uninstall preview/apply, and observation. `internal/runtime/noop` disables those capabilities.
- `web` is a separately deployed React SPA. It uses `/api/v1` by default or compile-time mock responses when explicitly enabled.

The backend uses external Go modules (SQLite driver, Argon2 support, and UUIDs); it is not standard-library-only.

## Request and operation flow

The browser bootstraps or logs into the single-admin session, then includes the CSRF token on mutations. Catalog detail and install-plan calls read the local catalog. Install validates configuration/conflicts, reserves generated ports within the process, writes app/revision/operation records, renders private files, and starts an operation goroutine. Lifecycle and uninstall follow the same persisted-operation pattern. Clients poll `GET /api/v1/operations/{id}`.

Workers are not an external queue. They are canceled on shutdown; after an unclean restart, persisted running operations are marked failed. There is no rollback or cross-instance lock. A successful uninstall removes the app record but intentionally leaves generated files and named volumes.

App reads invoke Compose observation and map service/container health and published ports into the response. This improves current-state reporting but is neither periodic monitoring nor event ingestion.

## Capability status

| Area | Status |
| --- | --- |
| SQLite persistence, migrations, demo seeding | Implemented |
| Single-admin auth, sessions, CSRF, direct-peer login limiting | Implemented |
| Local catalog, manifest/runtime validation, install review | Implemented |
| Docker Compose install/start/stop/restart/logs/uninstall/observation | Implemented, opt-out by configuration |
| Data retention on uninstall | Implemented as Compose down without volume deletion; recovery is manual |
| Asynchronous operations | Implemented in-process; persistence/restart handling is partial |
| Real frontend and explicit mock mode | Implemented |
| Container/systemd examples | Partial; examples disable Docker and SPA serving is separate |
| GitHub/remote catalog sources | Not implemented |
| Updates or update diffs | Not implemented |
| Backup, restore, or verification | Not implemented |
| Image-signature verification | Not implemented |
| Isolated privileged helper | Not implemented; checked-in unit is a future placeholder |
| Production/public-internet hardening | Not provided |

See the historical ADRs for original decisions; where they describe a read-only milestone, this document records the superseding implementation state.
