# Configuration reference

OpenDash reads environment variables at startup and exits on invalid values.

| Variable | Default | Meaning |
| --- | --- | --- |
| `OPENDASH_HTTP_HOST` | `127.0.0.1` | API listen host. Avoid public interfaces. |
| `OPENDASH_HTTP_PORT` | `8080` | API port, 1–65535. |
| `OPENDASH_LOG_LEVEL` | `info` | JSON log threshold: `debug`, `warn`, `error`, or `info` fallback. |
| `OPENDASH_READ_TIMEOUT` | `10s` | HTTP server read timeout (Go duration). |
| `OPENDASH_WRITE_TIMEOUT` | `10s` | HTTP server write timeout (Go duration). Long synchronous responses must fit. |
| `OPENDASH_SHUTDOWN_TIMEOUT` | `10s` | Graceful HTTP shutdown timeout (Go duration). |
| `OPENDASH_REQUEST_BODY_LIMIT` | `1048576` | Maximum request bytes; positive decimal integer. |
| `OPENDASH_USE_DOCKER_RUNTIME` | `true` | Use Docker CLI/Compose runtime. `false` selects a no-op runtime. |
| `OPENDASH_CATALOG_ROOT` | `./catalog` | Local directory containing `index.json` and manifests. |
| `OPENDASH_APPS_ROOT` | `./data/apps` | Managed root; projects install under `apps/<app-id>`. Must not be a symlink. |
| `OPENDASH_DB_PATH` | `./data/opendash.db` | SQLite database path; parent directory is created. |
| `OPENDASH_DEMO_MODE` | `false` | Seed demo dashboard data into an empty database. |
| `OPENDASH_AUTH_ENABLED` | `true` | Require sessions. When false, protected routes are open and auth actions return 404 except bootstrap status. Development only. |
| `OPENDASH_SECURE_COOKIES` | `false` | Set the session cookie `Secure` attribute. Enable when browser access is HTTPS. |
| `OPENDASH_SESSION_LIFETIME` | `24h` | Session duration from `1m` through `720h`. |

`IdleTimeout` is fixed at 120 seconds and is not currently configurable. Boolean values use Go's accepted boolean syntax (normally `true` or `false`). Relative paths resolve from the process working directory; service deployments should use absolute paths.

## Frontend

| Variable | Default | Meaning |
| --- | --- | --- |
| `VITE_USE_MOCK_API` | unset/false | Set exactly `true` at Vite build/dev time to use in-browser mock responses. |

Real mode calls `/api/v1` on the frontend origin. Vite development proxies `/api` to `http://localhost:8080`; production hosting must route that path to the API. The backend does not serve the SPA.

## Data and generated files

SQLite enables WAL, foreign keys, a busy timeout, transactional embedded migrations, and one pooled connection. Runtime install files are created below `OPENDASH_APPS_ROOT` with private directory/file permissions. Values can include secrets; protect both roots and their backups. Uninstall does not delete project files or named Docker volumes.
