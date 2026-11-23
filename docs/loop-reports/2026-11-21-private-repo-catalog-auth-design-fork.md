# Loop Report: 2026-11-21 - Private Repository Catalog Auth Design Fork

## What was chosen

The top-ranked candidate from `PLAN.md`'s "Later" section: **Expand catalog source support to private repositories (requires PAT/auth)**.

This was identified as a genuine design fork per the heuristics in `loop.md`:
- There are two defensible shapes for private repo catalog auth support already existing on branches:
  - `opendash/loop-2026-10-03-private-repo-catalog-auth-design-fork`
  - `opendash/loop-2026-11-21-private-repo-design-fork`
- The change involves auth infrastructure (PAT requirements, credential handling) with no immediate user-facing slice
- It cannot be verified by branch CI in the current state (schema change on trusted input boundary)

Per `loop.md`'s design fork rule: "If the choice is a genuine design fork (contradicts the threat model in `SECURITY.md` / ADRs, two defensible shapes), present the options and wait — don't guess past it."

## What changed

- **Created branch**: `opendash/loop-2026-11-21-private-repo-catalog-auth-design-fork` off `main` at commit `b5a26d3bda2dfec953cfd21c148826e3854c7a14` (date: 2026-11-21)
- **Created report**: `docs/loop-reports/2026-11-21-private-repo-catalog-auth-design-fork.md`
- **No code changes** — this is a design documentation task only, per the rule against shipping dead code for unverifiable schema changes

## What could not verify

- No user-facing slice to test — the change is purely infra/auth-layer
- CI cannot verify schema changes on the catalog trust boundary without actual private repository access
- Two existing design fork branches present conflicting approaches; no clear "correct" shape without human decision
- PAT-based authentication flow requires secrets infrastructure that is out of scope for this iteration

## What remains

- A human should review the two existing design fork branches (`opendash/loop-2026-10-03-private-repo-catalog-auth-design-fork` and `opendash/loop-2026-11-21-private-repo-design-fork`) and choose one coherent path forward
- Once the design shape is settled, implement the actual catalog source auth changes with proper CI verification
- Add PAT credential handling to the catalog source loading pipeline