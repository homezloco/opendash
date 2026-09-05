# Agent Notes

Last validated: current session.

## Scope and environment

- Go 1.22.2, Node v26.8.1/npm 11.19.0, and Docker 29.6.1 are available in the current environment.
- Respect the active task's file scope. Do not edit generated artifacts or dependencies unless explicitly requested.
- The backend uses SQLite, single-administrator authentication, a local catalog, and an implemented Docker Compose runtime. Docker access remains explicit and security-sensitive.
- The frontend uses the real API by default; set `VITE_USE_MOCK_API=true` only for UI-only development.

## Verified commands

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build -o opendash ./cmd/opendash
cd web && npm run typecheck && npm run test && npm run build
node validate-schemas.mjs
docker build -t opendash:latest .
```

Equivalent Make targets are `fmt`, `go-test`, `go-vet`, `go-build`, `web-typecheck`, `web-test`, `web-build`, `validate-schemas`, and `docker-build`.

Top-level orchestration targets:

- `make build` — build the SPA and the Go binary.
- `make test` — run Go tests and web tests.
- `make vet` — run `go vet ./...`.
- `make security` — run `go vet`, a `-trimpath` build, Docker socket permission checks, `npm audit --audit-level=high`, and `gosec` if installed.
- `make dist` — produce `dist/opendash-<version>-<os>-<arch>.tar.gz` containing the binary, built `web/dist`, systemd units, and `README.md`.
- `scripts/e2e-smoke.sh` — start the server on a temporary SQLite database, bootstrap the admin, list the catalog, and shut down cleanly (no Docker required).

The optional `docker_live` test mutates the Docker daemon and uses the exact resource prefix `opendash-test-hello`; follow `CONTRIBUTING.md` and verify cleanup. Never broaden socket permissions (for example with `chmod 666`).

## Current limitations

Remote/GitHub catalogs, destructive restore, configured isolated restore verification, recovery apply automation, image-signature verification, and privileged-helper isolation are not implemented. Backup archive creation/preview and recovery inventory export/import preview are implemented. Do not infer backend capabilities from frontend mock data or describe the application as production-ready.
