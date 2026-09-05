# User guide

## Bootstrap and sign in

On first launch, create the single administrator. The username must be 3–64 characters and the password 12–1024 characters. Bootstrap is one-time. A successful login invalidates the prior administrator session; there is no multi-user management or password-reset workflow.

The browser keeps an `HttpOnly` session cookie and the frontend sends the session CSRF token for mutations. If HTTPS is in use, enable secure cookies.

## Dashboard preferences

On the **Apps** page, open **Dashboard settings** to choose grid or list view, tile density and size, grouping and sorting, and which compact fields are hidden. Star favorite apps and use **Move to top** to define manual order. Preferences are saved in SQLite for the administrator and restored on later visits.

## Browse and install from the catalog

The catalog is loaded from the local `index.json`; OpenDash does not fetch GitHub or other remote catalogs. Open an app, request its install plan, and review:

- image references and digest-pinning warnings;
- generated localhost ports and named volumes;
- effective configuration (secret fields are marked sensitive);
- requested permissions and detected privileged/host-network/root/capability risks;
- existing app, project, path, and known fixed-port conflicts.

Review is advisory, not signature or provenance verification. Submit install only after resolving required configuration. Installation is asynchronous: the API records and returns an operation, while the frontend polls it. A second concurrent operation for the app is rejected.

## Backups and recovery

Manifests may declaratively identify named volumes to back up, whether configuration metadata is included, and migration/rollback risk. OpenDash never accepts executable backup hooks. The Backups page creates timestamped archives under `OPENDASH_BACKUPS_ROOT` (default `./data/backups`) using a temporary, network-disabled Docker container and applies count/age retention. Restore preview checks the archive checksum and displays volumes, images, manifest compatibility, risk, and current source trust without modifying data.

The Recovery page exports a JSON inventory of installed apps, manifest sources, and backup records. Import preview validates an inventory, reports ID conflicts, and highlights untrusted sources. Destructive restore and recovery apply remain disabled. Isolated verification is represented by the API but returns unavailable unless a safe drill verifier is configured; it must never overwrite the running application's volumes.

## Lifecycle and observation

Installed-app details provide start, stop, restart, and logs. These lifecycle actions also return asynchronous operations. Logs are a bounded Compose tail (100 lines by default; API accepts 1–10,000), not streaming or durable log storage.

When app data is read, OpenDash runs `docker compose ps` and updates displayed service state, health, and published endpoint ports. This is request-time observation, not continuous monitoring. Docker/Compose errors can leave persisted state stale or report an operation failure.

Operations execute in the API process. Graceful shutdown cancels active work; after a crash/restart, records left running are marked failed. No rollback, resumable queue, or multi-instance coordination is provided.

## Uninstall and retained data

Always review the uninstall plan. It lists Compose-derived containers, images, networks, and volumes and reports that data is retained. Applying uninstall runs `docker compose down` **without** `--volumes`:

- project containers and the Compose network are normally removed;
- named volumes remain;
- images are not removed;
- generated Compose and environment files remain below the apps root;
- the installed-app record is removed after successful completion.

Retention prevents automatic volume deletion, but it is not a backup, restore workflow, or recovery guarantee. Record the project/volume names before uninstall and manage retained resources manually with Docker. OpenDash has no reinstall/adopt or secure-delete workflow.

## GitHub manifest sources

The **Sources** page lets you add an OpenDash manifest hosted in a public GitHub repository. Paste the full `https://github.com/owner/repo/blob/ref/path/to/app.json` URL. OpenDash:

- accepts only `https://github.com` `/blob` or `/tree` URLs pointing to a `.json` manifest;
- rejects credentials, fragments, query strings, non-GitHub hosts, ambiguous paths, and private/localhost redirects;
- resolves the branch or tag to an immutable commit SHA;
- validates the manifest, stores its exact content in SQLite, and computes a checksum;
- marks direct sources as **untrusted** until you preview and explicitly confirm them, after which they become **user-trusted**.

An optional `OPENDASH_GITHUB_TOKEN` environment variable may be set server-side for higher rate limits. No token is ever accepted from the browser or stored in the database.

## Update previews

For apps installed from a GitHub source, the **Update preview** action resolves the latest commit at the same ref and compares it to the pinned version. The preview shows:

- manifest, configuration, Compose, image, port, and volume changes;
- a permission diff with added/removed/changed permissions;
- migration and rollback disclosures.

A moving branch/tag resolving to a new commit during preview is expected update discovery and does not lower trust. Different content returned for the same immutable commit is integrity drift: the source becomes **changed** and update is blocked.

To apply, confirm the exact preview commit and checksum. Additional acknowledgement is required for increased permissions, removed volumes, migration risk, or rollback that is not guaranteed. OpenDash preserves the prior generated revision, app ID, Compose project, configuration, and retained data volumes. It persists the new revision and source snapshot only after a health-gated Compose start succeeds. If startup fails it re-applies the prior Compose config only when the new manifest declares `upgrade.rollbackSafe: true`; this never rolls back database or volume contents. Declarative `upgrade.migrationRisk`, `upgrade.rollbackSafe`, and `upgrade.rollbackNotes` metadata is supported, but arbitrary hooks are not.

## Feature boundaries

Remote GitHub sources support preview, add, list, remove, and health-gated update application. Backups, database rollback, restore verification, recovery automation, image-signature checks, notifications, and continuous monitoring are not implemented. Protection/backup-looking rows in demo or mock data are illustrative dashboard data only.
