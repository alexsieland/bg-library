.PHONY: up down down-clean help build build-docker clean

export GHCR_USERNAME ?= alexsieland

help:
	@echo "Available targets:"
	@echo "  make up          - Build images (backend, frontend, database) and start the containers"
	@echo "  make down        - Stop the containers"
	@echo "  make build       - Build all Docker images (backend, frontend, database) without starting containers"
	@echo "  make clean       - Clean docker volumes (will stop containers if they are running)"
	@echo "  make help        - display this message"

up: build
	@echo "Bringing up containers..."
	@docker compose up -d

down:
	@echo "Stopping containers..."
	@docker compose down

build:
	@echo "Running build-docker in backend/..."
	@$(MAKE) -C backend GHCR_USERNAME=$(GHCR_USERNAME) build-docker
	@echo "Running build-docker in frontend/..."
	@$(MAKE) -C frontend GHCR_USERNAME=$(GHCR_USERNAME) build-docker

build-docker: build

clean:
	@echo "Stopping containers and removing volumes..."
	@docker compose down --volumes

