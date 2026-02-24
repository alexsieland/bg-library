# Backend - bg-library

This directory contains the Go backend for the board game library system.

## Development Setup

### Prerequisites

- Go 1.25+
- Docker (for the database)
- [sqlc](https://sqlc.dev/) (optional, if generating code)
- [oapi-codegen](https://github.com/deepmap/oapi-codegen) (optional, if generating code)

### Running the Database

The database is intended to be run via Docker for development. You can use the `docker-compose.yaml` in the project root to start only the database:

```bash
docker compose up -d db
```

### Makefile Commands

A `Makefile` is provided in this directory to simplify common tasks:

- `make generate`: Generates Go code from the SQL schema/queries (using `sqlc`) and OpenAPI specification (using `oapi-codegen`).
- `make build-docker`: Builds the local Docker image for the backend.
- `make run`: Runs the API server locally directly using `go run main.go`. Make sure the database is running and accessible.
- `make remove-docker`: Removes the local backend Docker image.
- `make format`: Formats and simplifies code using `gofmt` and `goimports`.
- `make lint`: Checks code formatting (excluding generated files).
- `make test`: Runs all unit tests.
- `make help`: Displays available targets.

### Running Unit Tests

To run all unit tests for the backend:

```bash
make test
```

To run a specific test, you can pass the `TEST_ARGS` variable:

```bash
make test TEST_ARGS="-run TestAddGame"
```

### Configuration

Environment variables can be configured in the `.env` file at the root of the project. The backend expects the following variables:

- `DB_PORT`
- `DB_HOST`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `GIN_MODE`
