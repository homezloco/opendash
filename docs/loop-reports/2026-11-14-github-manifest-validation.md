# Loop Report: GitHub Manifest Validation Support

## What I chose
Implement validation for GitHub-hosted manifests.

## Why
Per `PLAN.md`, this is the next step for external catalog integration.

## Changes
- Updated `internal/catalog/catalog.go`:
  - Added `gitURLPattern` to validate `manifest.Source` as a valid GitHub-hosted JSON file.
  - Updated `ValidateManifest` to check `manifest.Source` if provided.

## Verification
- Code review confirms the regex matches the requirements stated in `internal/githubsource/parse.go`.

## What remains
- Implementing the GitHub-hosted catalog resolution (which will consume these validated `Source` fields).
