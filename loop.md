# The improvement loop

How OpenDash improves itself. Each round takes the next-highest-leverage
increment, ships it verified and documented, and feeds what it learned
back into the next round. This file is the loop's durable state: the
invariants, the verification contract, and the ranking heuristics. When a
round teaches something the checklist missed, add it here.

## The loop

1. **Orient** — Read the "what remains" sources below and the latest
   report under `docs/loop-reports/` for candidates. If the operator
   named a specific item, that item wins.
2. **Choose** — Pick ONE increment, or one tight batch of related items,
   by the ranking heuristics below. If the choice is a genuine design
   fork (contradicts the threat model in `SECURITY.md` / ADRs, two
   defensible shapes), present the options and wait — don't guess past it.
3. **Check the blast radius first** — Before writing code, state what the
   change touches: the single-administrator auth/session/CSRF path, the
   Docker runtime boundary (`internal/runtime/docker`), the catalog trust
   boundary (manifests are trusted input), generated files under
   `OPENDASH_APPS_ROOT` (mode `0600` secrets), and SQLite schema/records
   (migrations are transactional; stored history is the asset). Also
   check second-order surfaces: could the change make the listener
   broader than loopback, broaden Docker socket access, or let frontend
   mock data masquerade as a real backend capability?
4. **Implement** — The smallest diff that satisfies the invariants.
   Reuse existing machinery before building new: the persisted-operation
   pattern for async work, `internal/catalog` validation for manifest
   changes, existing middleware for auth/CSRF. Follow `CONTRIBUTING.md`:
   `gofmt` for Go, no hand-editing generated output, contracts
   (`api/openapi.yaml`, schemas, fixtures) updated in the same commit.
5. **Verify** — The contract below, every time. New behavior gets new
   tests where the suite can express it.
6. **Document in the same commit** — `README.md`'s "Current scope" lists
   and `docs/architecture/current-state.md`'s capability table are
   load-bearing honesty; a change that alters capability status updates
   them in the same commit. `api/openapi.yaml` tracks route changes.
7. **Ship and report honestly** — One commit per feature, push on an
   `opendash/loop-*` branch, then report in
   `docs/loop-reports/<date>-<slug>.md`: what shipped, what was verified
   and how, and what's genuinely left — separated into real gaps, design
   forks, and diminishing returns. Never report "done" without the suite
   result that proves it, and never claim a capability the backend does
   not implement (mock data is not evidence).

## Invariants — never trade these away

- **Local-admin trust model.** OpenDash is pre-1.0, single-
  administrator, intended for trusted local administration. The default
  listener stays `127.0.0.1`. No round may add internet-facing
  assumptions, multi-tenant semantics, or weaken auth/CSRF/session
  handling.
- **Docker daemon safety.** Docker access is effectively root-equivalent
  and stays explicit and opt-out-able (`OPENDASH_USE_DOCKER_RUNTIME`).
  Never broaden socket permissions (no `chmod 666`), never invoke Docker
  in tests without the `docker_live` opt-in tag + env var, and keep
  managed paths constrained beneath `OPENDASH_APPS_ROOT` with the
  symlinked-root rejection intact.
- **Data retention on uninstall.** Uninstall intentionally leaves named
  volumes and generated project files. Do not silently change retention
  semantics in either direction — deleting retained data or claiming
  recovery automation that doesn't exist are both failures.
- **Honest capability reporting.** Catalog risk surfacing (unpinned
  digest, privileged mode, host networking, caps, root user) must stay
  accurate. Image-signature verification and privileged-helper isolation
  are *not implemented* — no change may describe them as shipped.
- **Secrets stay secret.** Manifest secret values land only in `0600`
  generated files and container environments — never in logs, API
  responses (beyond the documented marked-secret install-plan fields),
  or the frontend bundle.
- **SQLite records are the asset.** Migrations stay transactional;
  operation/app/revision history semantics don't subtly change.

## Verification contract

```bash
gofmt -l cmd internal            # must print nothing
go vet ./...
go test ./...
go build -o opendash ./cmd/opendash
cd web && npm ci && npm run typecheck && npm run test && npm run build
node validate-schemas.mjs
```

Equivalent Make targets: `fmt` (writes), `go-vet`, `go-test`,
`go-build`, `web-typecheck`, `web-test`, `web-build`,
`validate-schemas`.

- `go test ./...` does **not** run the live Docker test — it is gated
  behind `-tags=docker_live` and `OPENDASH_DOCKER_LIVE=1`, mutates a real
  daemon, and must never run in CI or from the loop. If a change could
  affect it, say so in the report and leave it to an authorized human.
- `scripts/e2e-smoke.sh` runs in CI (`e2e` job) — it uses a temp SQLite
  DB and no Docker. Changes to bootstrap/auth/catalog-listing should
  expect it to exercise them.
- `make security` (vet, trimpath build, docker.sock permission check,
  `npm audit`, gosec) mirrors `security.yml`; don't add a dependency
  without considering its audit posture.
- Re-read any large edit's seams — verify in the file, not the tool
  output.

## Ranking heuristics — what "next" means

1. **Security/threat-model gaps** — anything that weakens the
   invariants, or closes a real gap in `SECURITY.md`'s threat model,
   outranks everything.
2. **Correctness of implemented capabilities** — lifecycle/observation
   accuracy, persisted-operation recovery, catalog validation, CSRF/auth
   edges.
3. **Documented "not implemented" work, smallest useful slice** —
   `README.md` "Current scope" and `docs/architecture/current-state.md`
   capability table list remote-catalog gaps, destructive restore,
   update application, isolated privileged helper. A round may advance
   one honest slice; it may not paper over the gap with mock data.
4. **Contract drift** — `api/openapi.yaml`, JSON schemas/fixtures,
   `docs/` vs. actual behavior.
5. **Hygiene** — dead code, dependency rot, test coverage near existing
   seams.

## What remains — orientation sources

- `README.md` → "Current scope" (implemented / partial / not
  implemented) and "Verification".
- `docs/architecture/current-state.md` → capability-status table; ADRs
  for decisions a change must not contradict.
- `SECURITY.md` → trust and threat model; the authoritative list of what
  is deliberately not provided.
- `AGENTS.md` → verified commands and current limitations.
- `CONTRIBUTING.md` → development rules and the live-Docker-test
  procedure (humans only).
- `docs/loop-reports/` → prior rounds' honest leftovers.
- GitHub issues on `homezloco/opendash` → operator-filed work.

## The loop, hosted

This loop runs on the LiveGraph deployment: the "OpenDash Engineering"
graph ticks on a schedule, implements one increment via GitHub MCP on an
`opendash/loop-*` branch, CI on the branch (`.github/workflows/ci.yml`,
triggered on `opendash/**` pushes) is the verifier, and an approved
verdict routes to the gated "OpenDash Publisher" graph — the PR parks in
`/approvals` until a human ships it. The repo is the loop's memory:
reports live in `docs/loop-reports/` on each branch.

Honest limits: the engineer cannot run commands — Go/npm verification
happens in CI, not in the hop — and no CI job has an authorized Docker
daemon, so anything whose correctness depends on real Compose behavior
must be stated as an untested assumption in the report. The smallest-
diff rule matters more here, not less.
