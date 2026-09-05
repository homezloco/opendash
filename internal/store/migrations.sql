-- migration:1
CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY);
CREATE TABLE IF NOT EXISTS dashboard_data (kind TEXT PRIMARY KEY, payload BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS admins (id INTEGER PRIMARY KEY CHECK (id = 1), username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, created_at INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS sessions (token_hash BLOB PRIMARY KEY, admin_id INTEGER NOT NULL REFERENCES admins(id) ON DELETE CASCADE, csrf_token TEXT NOT NULL, expires_at INTEGER NOT NULL, created_at INTEGER NOT NULL);
CREATE INDEX IF NOT EXISTS sessions_expires_at ON sessions(expires_at);

-- migration:2
CREATE TABLE IF NOT EXISTS app_instances (
    id TEXT PRIMARY KEY,
    catalog_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    project_name TEXT NOT NULL UNIQUE,
    install_path TEXT NOT NULL UNIQUE,
    compose_path TEXT NOT NULL,
    status TEXT NOT NULL,
    health TEXT NOT NULL DEFAULT 'unknown',
    endpoints_json TEXT NOT NULL DEFAULT '[]',
    storage_json TEXT NOT NULL DEFAULT '[]',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS operations (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    output TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_operations_one_active_per_app
    ON operations(app_id) WHERE status IN ('pending', 'running');

CREATE TABLE IF NOT EXISTS revisions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    manifest_json BLOB NOT NULL,
    config_json BLOB NOT NULL,
    compose_path TEXT NOT NULL,
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS operations_app_id ON operations(app_id);

-- migration:3
CREATE TABLE IF NOT EXISTS manifest_sources (
    id TEXT PRIMARY KEY,
    source_url TEXT NOT NULL UNIQUE,
    owner TEXT NOT NULL,
    repo TEXT NOT NULL,
    path TEXT NOT NULL,
    ref TEXT NOT NULL,
    commit_sha TEXT NOT NULL,
    manifest_json BLOB NOT NULL,
    checksum TEXT NOT NULL,
    trust TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS manifest_sources_url ON manifest_sources(source_url);
CREATE INDEX IF NOT EXISTS manifest_sources_owner_repo ON manifest_sources(owner, repo);
ALTER TABLE app_instances ADD COLUMN source_id TEXT REFERENCES manifest_sources(id) ON DELETE SET NULL;

-- migration:4
CREATE TABLE backup_jobs (
    id TEXT PRIMARY KEY, app_id TEXT NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    status TEXT NOT NULL, error TEXT, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
);
CREATE INDEX backup_jobs_app_id ON backup_jobs(app_id);
CREATE TABLE backup_archives (
    id TEXT PRIMARY KEY, job_id TEXT NOT NULL REFERENCES backup_jobs(id) ON DELETE CASCADE,
    app_id TEXT NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE, path TEXT NOT NULL UNIQUE,
    size INTEGER NOT NULL, sha256 TEXT NOT NULL, manifest_sha256 TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL, verified_at INTEGER, integrity_ok INTEGER
);
CREATE INDEX backup_archives_app_created ON backup_archives(app_id, created_at DESC);
CREATE TABLE backup_schedules (
    id TEXT PRIMARY KEY, app_id TEXT NOT NULL UNIQUE REFERENCES app_instances(id) ON DELETE CASCADE,
    enabled INTEGER NOT NULL, interval_hours INTEGER NOT NULL, retention_count INTEGER NOT NULL,
    retention_days INTEGER NOT NULL, next_run_at INTEGER, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
);
