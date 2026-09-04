# ADR 002: Threat Model

## Status

Accepted and updated for the current mutating implementation. See `SECURITY.md` for operational guidance.

## Decision

1. **Docker privilege:** daemon access is effectively root-equivalent. The API directly invokes Docker/Compose; no isolated helper exists. Runtime use is optional and must be limited to a trusted account.
2. **Authentication and browser attacks:** a single Argon2id administrator uses opaque, hashed sessions, strict same-site HTTP-only cookies, and CSRF headers on authenticated mutations. Secure cookies require HTTPS.
3. **Network abuse:** authentication attempts are limited in memory by direct peer address; forwarded addresses are ignored. Proxies must add TLS, access policy, and their own rate limiting. Public exposure is not recommended.
4. **Information disclosure:** SQLite, generated environment/Compose files, Docker metadata, logs, and install-plan responses can contain sensitive configuration. Files are private but secrets are not encrypted or managed by a secret store.
5. **Catalog and command injection:** manifests are trusted local input, constrained by JSON Schema and runtime checks. Runtime commands use argument arrays, managed project-name patterns, and paths below a non-symlinked apps root. Risk review is advisory and does not sandbox containers.
6. **Resource exhaustion and availability:** body limits and HTTP/runtime timeouts are enforced. Operations execute asynchronously in-process; crashes mark interrupted records failed but provide no rollback or recovery.
7. **Supply chain:** dependencies and container images remain supply-chain inputs. Digest-pinning warnings exist, but OpenDash does not verify signatures, attestations, or provenance.
8. **Data lifecycle:** uninstall retains named volumes and generated files. This is neither backup verification nor secure deletion.

## Consequences

Run loopback-only or behind a trusted TLS proxy/firewall, keep authentication enabled, protect the database/apps roots, and never make a Docker socket world-writable. Treat the local catalog and anyone able to modify the OpenDash process as highly trusted. OpenDash is not production-ready and should not be internet-facing.
