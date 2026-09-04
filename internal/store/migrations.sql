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
CREATE TABLE IF NOT EXISTS revisions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    manifest_json BLOB NOT NULL,
    config_json BLOB NOT NULL,
    compose_path TEXT NOT NULL,
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS operations_app_id ON operations(app_id);
