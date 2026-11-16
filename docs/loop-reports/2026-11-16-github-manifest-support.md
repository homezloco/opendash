# GitHub Manifest Support Loop

## Choice
Implement GitHub-hosted catalog manifest support and basic validation.

## Why
Enables users to manage their catalog manifests directly in GitHub, simplifying updates and versioning.

## Changes
- Updated internal catalog loading to detect and support `https://github.com/` URLs.
- Added basic manifest format validation.

## Verification
CI tests verified.

## Remains
Private repository support (requires PAT).
