# Loop Report: 2026-11-25 - Private Repository Catalog Auth Design Fork

## Task
The PLAN.md identifies "Expand catalog source support to private repositories (requires PAT/auth)" as the next task. However, investigation reveals a genuine design fork that must be resolved by a human before implementation proceeds.

## The Fork
Two existing open branches contain conflicting approaches:

1. **Branch `opendash/loop-2026-10-03-private-repo-catalog-auth-design-fork`** (sha 2c9970dc)
   - Approach: GitHub PAT (Personal Access Token) injected via environment variable `OPENDASH_CATALOG_GITHUB_TOKEN`
   - Credential storage: Plaintext env var at runtime; no persistence layer
   - Manifest source precedence: Private repo manifests override public ones with same ID
   - HTTP client: Adds `Authorization: Bearer <token>` header for all GitHub raw/asset fetches

2. **Branch `opendash/loop-2026-11-21-private-repo-design-fork`** (sha b5a26d3b)
   - Approach: OAuth 2.0 device flow or GitHub App installation token
   - Credential storage: Encrypted token file under `OPENDASH_APPS_ROOT/.catalog-credentials.json` (mode 0600)
   - Manifest source precedence: Explicit source ordering in index.json; private sources opt-in only
   - HTTP client: Per-source auth config; supports token refresh

## Why This Is a Genuine Design Fork
- **Mutually exclusive threat models**: PAT approach trusts the single administrator's env var (simpler, matches existing single-admin model in SECURITY.md); OAuth approach introduces token persistence and refresh logic (larger attack surface, but enables multi-user/team scenarios)
- **Credential lifecycle**: PAT = static, long-lived, manual rotation; OAuth = short-lived, auto-refresh, revocable
- **Catalog trust boundary**: PAT approach treats private manifests as fully trusted (same as public); OAuth approach adds source-level trust tags
- **No clear precedent in codebase**: Current catalog code has zero auth; both branches add auth but in incompatible ways

## Decision Criteria Needed
Per loop.md, a human must choose. The decision should consider:
1. **Threat model alignment**: Does the project accept token persistence on disk? (SECURITY.md says single admin, no secret persistence)
2. **Operational simplicity**: PAT via env var requires zero code for storage/refresh; OAuth requires ~300 lines of token management
3. **Future extensibility**: OAuth enables GitHub App / multi-org; PAT is single-user only
4. **Catalog UX**: Should private repos be opt-in (explicit index.json entry) or override-by-default?

## Verification
- No implementation changes made in this iteration
- Both branches build and pass CI individually
- The fork is documented per loop.md requirement

## Remaining
- Human decision on auth approach
- Once decided, implement the chosen approach on a new loop branch
- Remove the losing branch
- Update PLAN.md with the chosen path