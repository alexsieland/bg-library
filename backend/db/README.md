# Database Container

PostgreSQL container with automatic schema and view initialization.

## Features

### Schema Initialization (`schema.sql`)
Tables are created automatically on first boot via `/docker-entrypoint-initdb.d/schema.sql`.

### Views (`views.sql`)
Database views are refreshed on **every** container start via `docker-entrypoint-views.sh`. This ensures views reflect the current schema even if the data volume is reused.

### Optional Data Import (`load-import.sh`)
Seed the database with a real-data export for local development or testing.

**How it works:**
- The import runs **only when the database volume is empty** (same guarantee as `schema.sql`)
- If no file is mounted at `/opt/db-import/import.sql`, the script exits silently
- Safe for CI and production — no import will run unless explicitly configured

**Usage:**
1. Export a database using `GET /api/v1/export/database`
2. Save the file locally (e.g., `./local-data/db_export_2026-05-27T14-25-30Z.sql`)
3. Add a volume mount to the `db` service in `compose.yaml`:
   ```yaml
   db:
     volumes:
       - db-data:/var/lib/postgresql
       - ./local-data/db_export_2026-05-27T14-25-30Z.sql:/opt/db-import/import.sql:ro
   ```
4. Run `docker compose up`

The import happens automatically on first boot.

**To reset and reimport:**
```bash
docker compose down -v  # Wipe the db-data volume
docker compose up       # Fresh import
```

## File Execution Order

Files in `/docker-entrypoint-initdb.d` run in sorted name order (digits before letters):

1. `01_schema.sql` — table definitions
2. `99_load-import.sh` — optional data import (if mounted)

Views are applied **after** initialization completes, on every boot.
