# syntax=docker/dockerfile:1

# Stage 1: Build
FROM golang:1.25.7 AS build

WORKDIR /usr/src/app

# Copy the dependency files first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the source code
COPY backend/ ./

# Compile the application
RUN go build -v -o /usr/local/bin/app ./

# Stage 2: Final Runtime
FROM debian:bookworm-slim

# Install CA certificates, pg_dump, Node.js, and npm
RUN apt-get update && apt-get install -y \
    ca-certificates \
    postgresql-client \
    nodejs \
    npm \
  && rm -rf /var/lib/apt/lists/*

# Pre-install pg-anonymizer at image build time
COPY backend/export/package.json /opt/pg-anonymizer/package.json
RUN npm install --prefix /opt/pg-anonymizer

# Copy the anonymizer config file
COPY backend/export/anonymizer-config.txt /opt/anonymizer-config.txt

WORKDIR /usr/src/app

# Set environment variable to indicate Docker environment
ENV IS_DOCKER=true

# Copy the compiled binary from the build stage
COPY --from=build /usr/local/bin/app /usr/local/bin/app

# Copy the swagger documentation
COPY swagger/ ./swagger/

EXPOSE 8080

CMD ["app"]
