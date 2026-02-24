# bg-library

A board game library management system.

## Project Structure

- `backend/`: Go backend source code and related tools.
- `swagger/`: API specification and UI.
- `docker-compose.yaml`: Docker Compose configuration for the entire stack.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Make

### Usage

Use the following `make` commands from the project root to manage the application:

- `make up`: Build the backend, create the Docker image, and start all containers (backend and database).
- `make down`: Stop and remove the containers.
- `make clean`: Stop the containers and remove the database volumes (useful for a fresh start).
- `make help`: Display available targets.

For more detailed information on the backend, see [backend/README.md](backend/README.md).

## Development

### Pre-commit Hook
To ensure consistent code formatting, a Git pre-commit hook is provided in the `scripts/` directory. This hook automatically runs `gofmt` and `goimports` on your staged files before each commit.

To install the hook, run the following command from the project root:

```bash
ln -s ../../scripts/pre-commit-hook.sh .git/hooks/pre-commit
```