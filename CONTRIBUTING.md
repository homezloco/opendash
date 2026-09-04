# Contributing to OpenDash

OpenDash welcomes focused bug fixes, documentation, tests, and features. It is a pre-1.0 project; describe behavioral and security consequences explicitly.

## Setup

Read `AGENTS.md`, `README.md`, `SECURITY.md`, and `CODE_OF_CONDUCT.md`. Install Go 1.22+, Node/npm, and optionally Docker with Compose v2. Then run:

```bash
cd web && npm install
cd ..
make web-typecheck web-test web-build
make go-test go-vet go-build
make validate-schemas
```

`npm test` sets `VITE_USE_MOCK_API=true`. The development frontend uses the real API unless that variable is explicitly set. The Vite proxy expects the API at `http://localhost:8080`.

## Development rules

- Keep changes small and update docs, `api/openapi.yaml`, schemas, and fixtures when their contracts change.
- Do not hand-edit generated output (`web/dist/`, binaries, coverage, or dependency directories).
- Format Go with `gofmt`; run the relevant Make targets before submitting.
- Catalog manifests must pass both JSON Schema validation and runtime restrictions described in `docs/manifest-authoring.md`.
- Never commit credentials, generated `.env` files, databases, or real server logs.

## Optional live Docker test

Unit tests mock command execution. An authorized developer may separately run the opt-in test that pulls and starts nginx:

```bash
OPENDASH_DOCKER_LIVE=1 go test -tags=docker_live ./internal/runtime/docker -run '^TestLiveNginxLifecycle$' -v
```

This test creates Docker resources with the exact Compose project/resource prefix `opendash-test-hello` and attempts best-effort cleanup. It mutates the authorized Docker daemon, may pull an image, allocates a localhost port, and intentionally verifies data-retaining uninstall. If the process is interrupted, inspect and remove only leftover `opendash-test-hello` containers, networks, and volumes after confirming ownership. Never run it against a daemon you are not authorized to modify.

## Pull requests

State what changed, which checks ran, and any untested Docker/deployment assumptions. Do not claim production readiness, backup/recovery, signature verification, remote catalog support, or helper isolation unless the implementation and tests exist.
