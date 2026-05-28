#!/bin/sh
set -e

# Write ~/.pgpass so pg_dump (invoked by pg-anonymizer during database export)
# can authenticate without the password appearing in process arguments or Go code.
#
# pg_dump silently ignores .pgpass if it is not owned by the current user or
# if the permissions are not exactly 0600, so both must be set explicitly.
#
# For local development (non-Docker): configure a local ~/.pgpass file or set
# the PGPASSFILE environment variable to point at an appropriate credentials file.
if [ -n "$DB_PASSWORD" ]; then
  PGPASS_HOST="${DB_HOST:-localhost}"
  PGPASS_PORT="${DB_PORT:-5432}"
  PGPASS_DB="${DB_NAME:-librarydb}"
  PGPASS_USER="${DB_USER:-postgres}"
  printf '%s:%s:%s:%s:%s\n' \
    "$PGPASS_HOST" "$PGPASS_PORT" "$PGPASS_DB" "$PGPASS_USER" "$DB_PASSWORD" \
    > /root/.pgpass
  chmod 0600 /root/.pgpass
fi

exec "$@"

