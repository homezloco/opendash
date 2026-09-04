# Manifest authoring

The current catalog is local and trusted. Each `catalog/index.json` entry points to `catalog/apps/<id>/app.json`. There is no GitHub fetch, remote refresh, signature verification, or arbitrary Compose-schema support.

Validate catalog files with:

```bash
node validate-schemas.mjs
```

This validates `schemas/app-manifest.schema.json` and `schemas/catalog-index.schema.json`. Runtime validation adds restrictions noted below; schema success alone is insufficient.

## Top-level manifest

Required fields are `id`, `name`, `version`, `description`, `category`, and `icon`. Unknown top-level fields are rejected by the schema.

- `id`: 1–64 lowercase alphanumeric kebab-case characters/segments.
- `name`: 1–128 characters; `description`: 1–512; `category`: 1–64; `icon`: 1–32.
- `version`: `major.minor.patch`, with the schema's optional alphanumeric/dot prerelease and build forms.
- Optional `tags` are unique 1–32 character strings. `website` and `source` must be HTTP(S) URIs. `maintainer` is 1–128 and `license` 1–64 characters.
- `compose` is optional in JSON Schema, but runtime install requires exactly one of `compose.file` or `compose.inline`.

## Compose

`compose` accepts only `file`, `inline`, `projectName`, and `mainService`.

A referenced `file` must match the schema's yml/yaml/json path pattern. Current runtime file loading uses that value relative to the OpenDash process working directory and copies it verbatim; it is not resolved relative to the manifest. Prefer inline manifests until deployment paths are controlled.

Inline Compose requires at least one service. Each service requires `image`; only these keys are accepted: `container_name`, `restart`, `ports`, `volumes`, `environment` (string values), `cap_add`, `cap_drop`, `privileged`, `network_mode`, `user`, `depends_on`, and restricted `healthcheck` fields (`test`, `interval`, `timeout`, `retries`, `disable`). Top-level inline keys are only `services`, `networks`, and `volumes`. Long-form Compose fields and other Compose keys are rejected.

Port strings support a container port, `host:container`, or IPv4-address/host/container form accepted by the runtime parser; ports must be 1–65535. Runtime validation requires service images and valid mappings. `mainService` selects the service receiving declared endpoints and storage; if omitted, selection from the service map is not deterministic. Set it explicitly.

The runtime requires the effective project name to start with `opendash-` and contain only letters, numbers, `_`, or `-`. Although the schema permits a broader `projectName`, other values fail at execution. Omit it to get `opendash-<id>`.

## Endpoints and storage

`endpoints` contains one or more `{label, port, kind, path?}` objects. `kind` is `web`, `api`, `tcp`, or `udp`; `path`, if present, begins `/`. Endpoint container ports must be unique at runtime. OpenDash allocates missing host ports on `127.0.0.1` and applies mappings to `mainService`. Allocation avoids concurrent OpenDash reservations but cannot eliminate races with other processes.

`storage` entries require unique runtime `name` values and absolute container `path` values; optional `defaultSizeGb` is positive but currently informational. OpenDash creates named volumes like `opendash-<id>-<lowercase-name>` and attaches them to `mainService`. Uninstall retains named volumes.

## Configuration, secrets, and permissions

`config` entries require a unique uppercase environment `key` matching `^[A-Z_][A-Z0-9_]*$`. Optional fields are `label`, `description`, `defaultValue` (string), `required`, and `secret`. Required values must be nonblank. Install overrides use the request's `config` map; declared config is injected into every inline service and written to the generated environment file.

`secrets` entries require an uppercase `name`, with optional `description` and `required`. The current implementation obtains secret values from the same config override map, but only declared `config` keys are retained by config application. Therefore declare a matching `config` entry with `secret: true` for usable secret input. Files are mode `0600`, not encrypted or backed by a secret manager.

`permissions` entries require `kind`: `network`, `privileged`, `capability`, `device`, `host-path`, `ipc`, `pid`, or `root-user`; `description` and `required` are optional. They are review metadata, not grants or enforcement. Runtime risk review additionally detects privileged services, host networking, selected high-risk capabilities, explicit root users, and images not pinned with `@sha256:`. Digest pinning is checked syntactically; no signature is verified.

Runtime validation also rejects duplicate config keys, storage names, and endpoint ports. Keep catalog index IDs, manifest IDs, and paths consistent and run both schema and Go tests before submission.
