#!/usr/bin/env bash
# e2e-smoke.sh
# Starts the OpenDash server on a temporary SQLite database, bootstraps the
# administrator, lists the catalog, and exits cleanly. This script is safe to run
# without Docker: it uses the no-op runtime and does not install real apps.
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

BIN="${1:-./opendash}"
TMPDIR="$(mktemp -d)"
trap 'rc=$?; if [ -f "$TMPDIR/pid" ]; then kill "$(cat "$TMPDIR/pid")" 2>/dev/null || true; fi; rm -rf "$TMPDIR"; exit $rc' EXIT

export OPENDASH_HTTP_PORT="${OPENDASH_HTTP_PORT:-18080}"
export OPENDASH_DB_PATH="$TMPDIR/opendash.db"
export OPENDASH_APPS_ROOT="$TMPDIR/apps"
export OPENDASH_BACKUPS_ROOT="$TMPDIR/backups"
export OPENDASH_USE_DOCKER_RUNTIME=false
export OPENDASH_CATALOG_ROOT="$PROJECT_ROOT/catalog"
export OPENDASH_WEB_ROOT="$PROJECT_ROOT/web/dist"

if [ ! -x "$BIN" ]; then
  echo "Building $BIN..."
  go build -o "$BIN" ./cmd/opendash
fi

"$BIN" > "$TMPDIR/server.log" 2>&1 &
echo $! > "$TMPDIR/pid"

HEALTH_URL="http://127.0.0.1:$OPENDASH_HTTP_PORT/api/v1/health"
for i in $(seq 1 60); do
  if curl -fs "$HEALTH_URL" >/dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "Server failed to start" >&2
    cat "$TMPDIR/server.log" >&2 || true
    exit 1
  fi
  sleep 0.5
done

COOKIE_JAR="$TMPDIR/cookies.txt"
API="http://127.0.0.1:$OPENDASH_HTTP_PORT/api/v1"

echo "Bootstrapping admin..."
curl -fs -c "$COOKIE_JAR" -b "$COOKIE_JAR" -X POST "$API/bootstrap" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"e2e-smoke-test-password"}' >/dev/null

echo "Listing catalog..."
curl -fs -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$API/catalog" >/dev/null

echo "Smoke test passed."
