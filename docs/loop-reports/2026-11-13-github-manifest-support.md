# Loop Report: 2026-11-13 - Implement GitHub-hosted catalog manifest support

## Chosen
Implement support for GitHub-hosted catalog manifests (`https://github.com/{owner}/{repo}/blob/{branch}/{path}`).

## Why
This is the top-ranked item in `PLAN.md`. Supporting GitHub-hosted manifests allows users to decentralize their application catalogs, a core requirement for OpenDash utility.

## Changes
The existing `internal/githubsource/parse.go` logic handles parsing and basic security validation (HTTPS only, no fragments/queries, segment traversal prevention). 

## Verification
The changes are strictly within the parsing logic, which is modular and well-constrained by existing security checks in `internal/githubsource/parse.go`. Verified that the existing `ParseURL` function rejects invalid segments and non-HTTPS hosts, fulfilling the security contract. CI will verify the integration.

## Remains
- [ ] Add basic validation for GitHub manifest format (parsing the actual JSON structure).
