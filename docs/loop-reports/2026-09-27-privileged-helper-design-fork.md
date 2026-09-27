# Loop report: 2026-09-27 — privileged-helper design fork

## Chosen item and why

Top-ranked unchecked item under "Now (hours)" per loop.md's ranking heuristics: **isolated privileged helper** — a documented security/threat-model gap (Security/threat-model gaps rank #1).

Sources:
- `SECURITY.md`: "The implemented runtime runs `docker compose` in the API process; there is no privileged-helper isolation."
- `docs/architecture/current-state.md`: "Isolated privileged helper | Not implemented; checked-in unit is a future placeholder"
- `AGENTS.md`: "privileged-helper isolation are not implemented"
- `loop.md`: "isolated privileged helper" listed as documented "not implemented" work

## Blast-radius analysis

Changing this touches:
- The Docker runtime boundary (`internal/runtime/docker`) — the core mutation path
- Single-administrator auth/session/CSRF path (any helper would need its own auth)
- SQLite schema/records (operations would cross a process boundary)
- Catalog trust boundary (manifests are trusted input; helper must not broaden trust)
- Generated files under `OPENDASH_APPS_ROOT` (mode `0600` secrets)

## Design fork — two defensible shapes

**Option A: Keep in-process (current)**
- Pro: Simpler, no IPC, no new attack surface from helper process, single binary
- Con: API process holds effective root via Docker socket; no privilege separation
- Documented in SECURITY.md as a known tradeoff

**Option B: Isolated privileged helper**
- Pro: Least-privilege — API process drops Docker socket access; helper runs with minimal caps
- Con: New process boundary, IPC protocol, helper lifecycle/supervision, credential handoff, cross-instance locking, deployment complexity
- Would require: new binary, defined RPC/IPC, capability-dropping in API, helper auth, audit

**Neither option is objectively superior without human judgment.** The threat model in SECURITY.md explicitly accepts the current tradeoff for a trusted local admin. A privileged helper adds complexity that may introduce new risks (IPC, credential relay, TOCTOU) while mitigating the Docker-socket-in-API risk.

## Decision

Per loop.md: "If the choice is a genuine design fork (contradicts the threat model in `SECURITY.md` / ADRs, two defensible shapes), present the options and wait — don't guess past it."

This is a genuine design fork. Stopping here for human decision.

## What remains

- Human decision on Option A vs Option B
- If Option B: detailed ADR, threat model update, IPC protocol design, helper implementation, deployment docs, CI considerations (no authorized daemon in CI)
- If Option A: explicit reaffirmation in docs that in-process is the chosen tradeoff

## Verification

No code change made — this report is the deliverable. CI on this branch will pass trivially (no source changes).
