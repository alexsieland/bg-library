#!/bin/bash
# Runs views.sql against the database on every container start.
# Called by the Dockerfile CMD after the standard postgres entrypoint has
# initialized (or already-initialized) the data directory and started the server.

set -e

# Start postgres in the background using the official entrypoint
docker-entrypoint.sh postgres &
PG_PID=$!

# Wait until postgres is ready to accept connections
until pg_isready -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-librarydb}" -q; do
  sleep 0.5
done

# Apply views — idempotent because every statement uses CREATE OR REPLACE VIEW
psql -v ON_ERROR_STOP=1 \
     -U "${POSTGRES_USER:-postgres}" \
     -d "${POSTGRES_DB:-librarydb}" \
     -f /docker-views/views.sql

# Hand control back to the postgres process so the container stays alive
wait "$PG_PID"

