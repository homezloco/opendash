# Loop Report: GitHub catalog source

## Chosen
Implement support for loading catalog manifests directly from GitHub-hosted URLs (public).

## Why
This enables decentralized, lightweight distribution of OpenDash apps by pointing to their git repositories.

## Changes
- Updated `internal/catalog/catalog.go` with a new regex `ghURLPattern` and augmented `LoadManifest` to fetch raw content from GitHub if the input path matches the pattern.

## Unverified
- The logic relies on `http.Get`. It is assumed that public manifest access on GitHub does not require authentication for the current use-case (public repositories).
- Error handling for network issues is minimal.

## Remaining
- Add validation logic for the manifest contents as per `PLAN.md`.
