# Troubleshooting

## Vite reports `502 Bad Gateway` for `/api`

Real frontend mode proxies `/api` to `http://localhost:8080`. A 502 usually means the Go backend is not running, exited because configuration/database startup failed, or is listening on another host/port. Start it, inspect its JSON logs, and verify:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

Alternatively, for UI-only development restart Vite with `VITE_USE_MOCK_API=true npm run dev`. Do not use mock mode to validate backend behavior.

## Docker or Compose is unavailable

Run `docker version` and `docker compose version` as the same user running OpenDash. Check the readiness endpoint after authentication. If socket access is denied, use the daemon's restricted group or the package's supported Snap interface; never make the socket world-writable. The distroless OpenDash container does not include the Docker CLI, and `compose.dev.yml` intentionally disables runtime actions.

## Catalog is empty or fails to load

Confirm `OPENDASH_CATALOG_ROOT` points to a directory containing `index.json`, and that each entry's manifest exists beneath its `apps` directory. Relative paths are relative to the API working directory. Run `node validate-schemas.mjs`, then `go test ./internal/catalog ./internal/api` for runtime restrictions not covered by JSON Schema.

## Install fails after review

Check the operation's `error`, API logs, daemon connectivity, image access, fixed-port conflicts, required config, and free space. Generated endpoint ports have a small race window against unrelated host processes. A failed install may leave a persisted app/revision and generated files; there is no automatic rollback. Inspect before manual cleanup.

## An operation remains running

The frontend polls persisted operation state. Refresh and query the operation endpoint. Graceful shutdown cancels workers; restart marks records left running as failed. Work cannot resume, and operations are not coordinated across multiple API instances.

## App status or endpoint is stale

App reads trigger `docker compose ps`; they do not continuously watch Docker. Verify the Compose project directly and reload the app. If Docker observation fails, persisted values can remain stale.

## Uninstalled data still consumes disk

This is expected. Uninstall omits `--volumes` and leaves generated project files and images. Use the uninstall preview/project names to identify resources, verify ownership, then manage them manually with Docker. OpenDash does not verify backups or provide restore/adopt/secure-delete workflows.

## Login or CSRF failures

A successful login invalidates older sessions. Refresh to obtain the current session/CSRF token. Behind HTTPS, set secure cookies; over plain HTTP, a secure cookie will not be sent. Rate limiting keys on the direct peer, so all users behind one reverse proxy share a bucket unless the proxy supplies its own controls.
