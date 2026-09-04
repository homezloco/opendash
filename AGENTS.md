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

Equivalent Make targets are `web-typecheck`, `web-test`, `web-build`, `go-test`, `go-vet`, `go-build`, `validate-schemas`, and `docker-build`.

The optional `docker_live` test mutates the Docker daemon and uses the exact resource prefix `opendash-test-hello`; follow `CONTRIBUTING.md` and verify cleanup. Never broaden socket permissions (for example with `chmod 666`).

## Current limitations

Remote/GitHub catalogs, update diffs, backup/restore verification, image-signature verification, recovery automation, and privileged-helper isolation are not implemented. Do not infer backend capabilities from frontend mock data or describe the application as production-ready.
