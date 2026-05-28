# syntax=docker/dockerfile:1

# Dockerfile for PostgreSQL with schema initialization
FROM postgres:18.2

# Copy the schema file into the initialization directory (runs on first boot only)
COPY backend/db/schema.sql /docker-entrypoint-initdb.d/

# Copy views and the wrapper entrypoint
COPY backend/db/views.sql /docker-views/views.sql
COPY backend/db/docker-entrypoint-views.sh /usr/local/bin/docker-entrypoint-views.sh
RUN chmod +x /usr/local/bin/docker-entrypoint-views.sh

# Override CMD: the wrapper starts postgres, applies views on every boot, then waits
CMD ["docker-entrypoint-views.sh"]
