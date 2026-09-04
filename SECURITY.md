# Security Policy

OpenDash can control Docker workloads and handle application configuration. It is pre-1.0, intended for a trusted local administrator, and should not be exposed directly to the public internet.

## Reporting

Do not open a public issue for a vulnerability. Email the maintainers or use the repository's private vulnerability reporting feature with affected versions and reproduction steps. Security fixes target the latest default-branch revision.

## Trust and threat model

- **Docker authority:** access to the Docker daemon is effectively root-equivalent on a typical host. The implemented runtime runs `docker compose` in the API process; there is no privileged-helper isolation. Keep `OPENDASH_USE_DOCKER_RUNTIME=false` unless needed, grant socket access only to a dedicated trusted account, and never make the socket world-writable.
- **Network exposure:** the default listener is `127.0.0.1`. Keep it loopback-only or behind an authenticated TLS reverse proxy/firewall on a trusted network. OpenDash does not implement proxy-aware client identity or a distributed edge rate limiter.
- **Authentication:** one administrator is stored in SQLite. Passwords use Argon2id; opaque session-token hashes are stored in SQLite. Login replaces existing administrator sessions. Cookies are `HttpOnly` and `SameSite=Strict`; set `OPENDASH_SECURE_COOKIES=true` only when clients use HTTPS.
- **Rate limiting:** bootstrap and login use an in-memory limiter keyed by the direct TCP peer. Forwarded-IP headers are deliberately ignored. A reverse proxy therefore appears as one peer; it must provide its own abuse controls. Limiter state resets on restart and is not shared across instances.
- **CSRF:** authenticated non-GET/HEAD/OPTIONS requests require the session's `csrfToken` in `X-CSRF-Token`. Same-site cookies and security headers add defense in depth. Disabling authentication removes this protection and is only suitable for tightly controlled development.
- **Secrets:** manifest config and secret values are written into generated Compose/environment files under `OPENDASH_APPS_ROOT` with mode `0600`, and may be passed into container environments. They are not encrypted, backed up, redacted from every Docker inspection surface, or managed by a secret store. Install-plan responses mark secret fields but should still be handled as sensitive.
- **Catalog and workloads:** the local catalog is trusted input. Install review reports selected risks (unpinned image digest, privileged mode, host networking, high-risk capabilities, root user, and declared permissions), but does not sandbox workloads, verify image signatures, or prove image provenance.
- **Availability:** request sizes and server/runtime command durations are bounded. Asynchronous operations run in-process; shutdown cancels them, and records left running after a crash are recovered as failed. This is not transactional rollback or disaster recovery.
- **Retention:** uninstall runs Compose `down` without `--volumes`, so named data volumes remain. Generated project files also remain. Retention is not a verified backup, restore, secure deletion, or automated recovery facility.

Security headers include `nosniff`, frame denial, a same-origin content policy, referrer policy, and restrictive permissions policy. Managed runtime paths are constrained beneath the configured apps root and reject a symlinked root.

See [Docker access](docs/getting-started.md#docker-and-snap-socket-access), [configuration](docs/configuration.md), and [ADR 002](docs/architecture/adr-002-threat-model.md).
