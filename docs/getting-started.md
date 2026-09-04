# Getting started

## Demo mode with the real frontend

```bash
go build -o opendash ./cmd/opendash
OPENDASH_DEMO_MODE=true OPENDASH_USE_DOCKER_RUNTIME=false ./opendash
# Separate terminal
cd web && npm install && npm run dev
```

Browse to <http://localhost:5173>. Bootstrap the sole administrator, then log in. Demo rows are inserted only when the SQLite database is empty. The no-op runtime reports Docker unavailable and refuses lifecycle mutation; it does not simulate containers.

For a disposable database, point `OPENDASH_DB_PATH` and `OPENDASH_APPS_ROOT` into a temporary directory. Existing demo rows are persistent and are not removed merely by turning demo mode off.

## Real mode

Install Docker Engine and Compose v2, ensure the same OS user can run `docker version` and `docker compose version`, then start the API with the repository catalog:

```bash
OPENDASH_DEMO_MODE=false \
OPENDASH_DB_PATH="$PWD/data/opendash.db" \
OPENDASH_CATALOG_ROOT="$PWD/catalog" \
OPENDASH_APPS_ROOT="$PWD/data/apps" \
OPENDASH_USE_DOCKER_RUNTIME=true \
./opendash
```

Run `cd web && npm run dev` separately. The frontend proxies `/api` to port 8080. Runtime-generated Compose and environment files are placed below the apps root; catalog installation can pull images and create containers, networks, and named volumes.

## Docker and Snap socket access

Docker daemon access is normally granted through a restricted Unix group. Add only the dedicated OpenDash service user to the daemon's owning group, re-login/restart the service, and verify access as that user. Group membership is effectively root-equivalent; limit who can alter the OpenDash binary, catalog, configuration, and service account.

Docker installed as a Snap may expose a Snap-managed socket or require Snap interface connections depending on packaging/version. Inspect the active daemon endpoint with `docker context inspect` and the socket ownership with `stat`, then configure the service user's group and `DOCKER_HOST` only if your installation requires it. Follow the Docker/Snap package documentation for that host. Do **not** use `chmod 666`/`chmod 777`, expose an unauthenticated TCP daemon, or mount a socket into an untrusted container.

The supplied distroless Docker image contains neither the Docker CLI nor a supported socket-enabled deployment. `compose.dev.yml` deliberately disables runtime mutation and mounts no socket. Use the native binary for current real-runtime testing.

## Deployment notes

`compose.dev.yml` is a hardened demo API example with persistent SQLite data and Docker disabled. The API image does not serve `web/dist`; deploy the SPA separately or use Vite only for development.

`deploy/systemd/opendash-api.service` is a starting point: it binds loopback, uses secure cookies, and disables Docker. Secure cookies require HTTPS at the browser-facing endpoint. If adapting systemd for Docker, account for daemon-group access and the existing namespace restrictions; the checked-in future runtime-helper unit has no corresponding executable and is not implemented.

OpenDash is not recommended for public exposure or production use.
