# Dockerfile for PostgreSQL with schema initialization
FROM postgres:18.2

# Copy the schema file into the initialization directory
COPY schema.sql /docker-entrypoint-initdb.d/

# The postgres image will automatically run any .sql files in
# /docker-entrypoint-initdb.d/ when the database is first initialized

