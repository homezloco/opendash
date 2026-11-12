# Loop Report: 2026-10-03-private-repo-catalog-auth-design-fork

## Chosen item and why

Item: **Expand catalog source support to private repositories (requires PAT/auth)** from `PLAN.md`.

Why:
- The items under `PLAN.md` "Now (hours)" (`Implement support for GitHub-hosted catalog manifests` and `Add basic validation for GitHub manifest format`) are both in flight on open pull requests (#10 and #9 respectively).
- Following `loop.md` queue rules, in-flight items are bypassed to select the next candidate in the queue: private repository catalog support.

## Blast-radius analysis

Changing this touches:
- Catalog manifest fetch and resolution logic (`internal/githubsource` / `internal/catalog`).
- Secrets and credential handling boundaries (`SECURITY.md` invariants).
- Configuration / SQLite persistence model if per-source credentials are store-backed.
- OpenAPI specification and UI components if credentials are submitted or surfaced via API.

## Design Fork — two defensible shapes

**Option A: Environment-variable-based authentication (`OPENDASH_GITHUB_TOKEN`)**
- *Pros*: Simple and lightweight, no SQLite schema or API contract changes required, zero database persistence of secrets, aligns cleanly with local single-admin runtime environment.
- *Cons*: Applies globally to all remote GitHub catalog sources; cannot scope different authentication tokens for different private repositories.

**Option B: Per-source credential persistence (Encrypted SQLite storage + API extension)**
- *Pros*: Enables granular access control with scoped tokens per catalog source repository.
- *Cons*: Requires secret storage and encryption in SQLite, API contract updates with write-only/redacted token fields, increases credential exposure surface, and requires explicit key lifecycle design.

## Decision

Per `loop.md`: *"If the choice is a genuine design fork (contradicts the threat model in SECURITY.md / ADRs, two defensible shapes), present the options and wait — don't guess past it."* Furthermore, `loop.md` strictly instructs to never add unapproved secret handling mechanisms without explicit operator alignment.

This is a genuine design fork around credential handling. Stopping here for human choice between Option A and Option B.

## What remains

- Human architectural decision between Option A (environment token) and Option B (per-source stored credentials).
- Implementation of selected authentication mechanism for remote catalog manifest retrieval once decided.

## Verification

- No application code changes made; this report presents the design decision for operator evaluation. CI on this branch will pass trivially.
