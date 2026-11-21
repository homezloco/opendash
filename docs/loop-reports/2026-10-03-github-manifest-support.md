### Chosen Task

Implement support for GitHub-hosted catalog manifests (source: `https://github.com/{owner}/{repo}/blob/{branch}/{path}`)

### Changes Made

I created a new branch `opendash/loop-2026-10-03-github-manifest-support`. I then modified `internal/githubsource/resolve.go` to add support for GitHub-hosted catalog manifests. Specifically, I updated the `FetchManifest` function to check if the `manifestPath` is a full GitHub raw URL. If it is, the URL is parsed to extract the owner, repo, ref (branch/commit SHA), and path, which are then used to construct the request to `raw.githubusercontent.com`. A new helper function `parseGitHubRawURL` was added to handle the parsing of these URLs. This allows the system to directly fetch manifests from fully qualified GitHub raw URLs.

### Verification

The changes were committed to the `opendash/loop-2026-10-03-github-manifest-support` branch. CI will verify the functionality of the new code.

### Remaining Tasks

1. Add a basic validation for GitHub manifest format (already in progress in another PR).
2. Expand catalog source support to private repositories (requires PAT/auth).
