# OpenDash

OpenDash is a pre-1.0 home-server dashboard with a Go API, SQLite persistence, single-administrator authentication, a local application catalog, Docker Compose lifecycle control, and a React frontend. It is intended for trusted, local administration and is not production-ready.

## Getting started

Requirements: Go 1.22+, Node/npm, and (for real lifecycle operations) Docker with Compose v2 and authorized daemon access.

### Demo data with the real API

```bash
go build -o opendash ./cmd/opendash
OPENDASH_DEMO_MODE=true OPENDASH_USE_DOCKER_RUNTIME=false ./opendash
# In another terminal:
cd web && npm install && npm run dev
```

Open <http://localhost:5173>, create the sole administrator (username 3–64 characters, password 12–1024 characters), and sign in. Demo mode seeds an empty SQLite database; disabling Docker makes lifecycle calls unavailable rather than simulated.

### Real local catalog and Docker runtime

```bash
go build -o opendash ./cmd/opendash
OPENDASH_DB_PATH="$PWD/data/opendash.db" \
OPENDASH_CATALOG_ROOT="$PWD/catalog" \
OPENDASH_APPS_ROOT="$PWD/data/apps" \
./opendash
# In another terminal:
cd web && npm install && npm run dev
```

The runtime invokes the local `docker` CLI and `docker compose`; the process user must be authorized to use the Docker daemon. Review [Docker access](docs/getting-started.md#docker-and-snap-socket-access) before enabling it. Catalog installs, start/stop/restart, log tails, runtime observation, and data-retaining uninstall are implemented.

For UI-only work, run `VITE_USE_MOCK_API=true npm run dev`. Mock data can describe features that the backend does not implement; it is not evidence of real backup or protection behavior.

## Documentation

- [Getting started and deployment](docs/getting-started.md)
- [Configuration reference](docs/configuration.md)
- [User guide](docs/user-guide.md)
- [Manifest authoring](docs/manifest-authoring.md)
- [Architecture and current state](docs/architecture/current-state.md)
- [Security and threat model](SECURITY.md)
- [Development and testing](CONTRIBUTING.md)
- [Troubleshooting](docs/troubleshooting.md)
- [OpenAPI description](api/openapi.yaml)

## Current scope

**Implemented:** SQLite/WAL persistence and migrations; optional demo seeding; single-admin sessions and CSRF checks; local filesystem catalog; install/uninstall review; asynchronous install and lifecycle operations; Docker Compose mutation, logs, and observation; retained named volumes on uninstall; real and mock frontend API modes.

**Partial:** operation records survive in SQLite, but work runs in-process and interrupted running operations are marked failed at restart; observation occurs while app responses are read, not by a continuous monitor; uninstall retains Docker volumes and project files but does not provide recovery automation; deployment examples default to a no-Docker runtime.

**Not implemented/planned work:** remote/GitHub catalog sources, update planning or diffs, backup/restore and verification, image-signature verification, recovery workflows, and an isolated privileged runtime helper. Do not treat OpenDash as an internet-facing or production-ready control plane.

## API and persistence

Routes are under `/api/v1`; see `api/openapi.yaml`. Authentication uses an opaque `HttpOnly`, `SameSite=Strict` cookie. Mutating authenticated requests require the session `csrfToken` in `X-CSRF-Token`. Login replaces the prior administrator session. Authentication rate limiting uses the direct peer address and ignores forwarded-IP headers.

SQLite uses WAL, foreign keys, a busy timeout, embedded transactional migrations, and one pooled connection. Demo rows are seeded only into an empty database when `OPENDASH_DEMO_MODE=true`; normal mode reports empty/offline data honestly.

## Verification

```bash
make web-typecheck web-test web-build
make go-test go-vet go-build
make validate-schemas
make docker-build
```

See [development and testing](CONTRIBUTING.md) for the optional destructive `docker_live` test.
