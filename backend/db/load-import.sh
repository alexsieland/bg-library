#!/usr/bin/env bash
# load-import.sh — optional data import for local development
#
# This script runs inside /docker-entrypoint-initdb.d and is therefore only
# executed when the database data directory is empty (i.e. a fresh volume).
# It will never run against an already-initialized database.
#
# To seed a local database with a real-data export:
#   1. Place your exported .sql file on the host.
#   2. Mount it into the container at /opt/db-import/import.sql via
#      compose.override.yaml (see the comment in compose.yaml for an example).
#   3. Run `docker compose up` — the import runs automatically on first boot.
#
# If no file is mounted at /opt/db-import/import.sql this script exits silently,
# making it safe for CI and production where no import is desired.

set -e

IMPORT_FILE="/opt/db-import/import.sql"

if [ -f "$IMPORT_FILE" ]; then
  echo "load-import: loading $IMPORT_FILE into $POSTGRES_DB..."
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$IMPORT_FILE"
  echo "load-import: import complete."
else
  echo "load-import: no file found at $IMPORT_FILE, skipping."
fi

