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

# Install CA certificates to make secure HTTPS calls (if needed)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src/app

# Set environment variable to indicate Docker environment
ENV IS_DOCKER=true

# Copy the compiled binary from the build stage
COPY --from=build /usr/local/bin/app /usr/local/bin/app

# Copy the swagger documentation
COPY swagger/ ./swagger/

EXPOSE 8080

CMD ["app"]
