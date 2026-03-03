# bg-library

A full-stack board game library management system designed for conventions and festivals.

## Key Features
- Board game search and management.
- Patron management
- Simple checkout and check-in workflow.

## Project Structure

- `frontend/`: Svelte frontend source code.
- `backend/`: Go backend source code and related tools.
- `swagger/`: API specification and UI.
- `docker-compose.yaml`: Docker Compose configuration for the entire stack.

## Project Documentation
- [API Documentation](swagger/index.html)
- [Project Overview](docs/project-overview.md)
- [Functional Requirements](docs/functional-requirements.md)
- [Coding Guidelines](docs/coding-guidelines.md)
- [Testing Guidelines](docs/testing-guidelines.md)

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

For more detailed information on the frontend, see [frontend/README.md](frontend/README.md).
For more detailed information on the backend, see [backend/README.md](backend/README.md).

## Docker Environment Variables

The application uses environment variables to configure Docker containers. See [compose.yaml](compose.yaml) for the default configuration and how to override these variables.

### Frontend Variables
- `API_URL` - The backend API endpoint (default: `http://localhost:8080`)
- `BACKEND_PORT` - The backend port (default: `8080`)
- `EXPOSE_SWAGGER_UI` - Whether to expose Swagger UI (default: `false`)
- `NGINX_HOST` - The nginx server name (default: `localhost`)
- `FRONTEND_PORT` - The frontend port (default: `80`)
- `TRUSTED_PROXIES` - Comma-separated list of trusted proxy IPs/CIDRs (optional, see [Proxy Configuration](docs/proxy-configuration.md))
- `REAL_IP_HEADER` - HTTP header containing real client IP (default: `X-Real-IP`, see [Proxy Configuration](docs/proxy-configuration.md))

### Backend Variables
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `GIN_MODE` - Gin framework mode (`release` or `debug`)
- `CORS_ALLOWED_ORIGIN` - CORS origin (default: `*`)
- `BACKEND_PORT` - Backend API port (default: `8080`)

### GHCR Configuration
The application uses GitHub Container Registry (ghcr.io) for Docker images. Set the `GHCR_USERNAME` environment variable to specify the image repository owner:

```bash
GHCR_USERNAME=your-username docker compose up
```

The `compose.yaml` file defaults to `alexsieland` if `GHCR_USERNAME` is not set.

## Development

### Pre-commit Hook
To ensure consistent code formatting, a Git pre-commit hook is provided in the `scripts/` directory. This hook automatically runs `gofmt` and `goimports` on your staged files before each commit.

To install the hook, run the following command from the project root:

```bash
ln -s ../../scripts/pre-commit-hook.sh .git/hooks/pre-commit
```