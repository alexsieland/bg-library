# syntax=docker/dockerfile:1

# Stage 1: Install pg-anonymizer dependencies
# npm is only required here to install pg-anonymizer; it is intentionally excluded
# from the final runtime image. Only the node_modules output is copied forward.
FROM node:lts-slim AS node-deps
WORKDIR /opt/pg-anonymizer
COPY backend/export/package.json .
RUN npm install --omit=dev

# Stage 2: Build
FROM golang:1.25.7 AS build

WORKDIR /usr/src/app

# Copy the dependency files first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the source code
COPY backend/ ./

# Compile the application
RUN go build -v -o /usr/local/bin/app ./

# Stage 3: Final Runtime
FROM debian:bookworm-slim

# Install CA certificates, pg_dump, and the Node.js runtime.
# npm is NOT installed here — it was only needed in the node-deps stage.
RUN apt-get update && apt-get install -y \
    ca-certificates \
    postgresql-client \
    nodejs \
  && rm -rf /var/lib/apt/lists/*

# Copy pre-built pg-anonymizer node_modules from the node-deps stage.
# No npm required at runtime; nodejs alone is sufficient to execute the tool.
COPY --from=node-deps /opt/pg-anonymizer/node_modules /opt/pg-anonymizer/node_modules

# Copy the anonymizer config file
COPY backend/export/anonymizer-config.txt /opt/anonymizer-config.txt

# Copy the entrypoint script that writes ~/.pgpass at container startup so
# pg_dump (invoked by pg-anonymizer) can authenticate without the password
# appearing in Go code or subprocess argument lists.
COPY backend/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /usr/src/app

# Set environment variable to indicate Docker environment
ENV IS_DOCKER=true

# Copy the compiled binary from the build stage
COPY --from=build /usr/local/bin/app /usr/local/bin/app

# Copy the swagger documentation
COPY swagger/ ./swagger/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["app"]
