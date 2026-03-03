.PHONY: up down down-clean help build clean

help:
	@echo "Available targets:"
	@echo "  make up          - Build and start the containers"
	@echo "  make down        - Stop the containers"
	@echo "  make build       - Build the Docker images without starting containers"
	@echo "  make clean       - Clean docker volumes (will stop containers if they are running)"
	@echo "  make help        - display this message"

up:
	@echo "Running generate and build-docker in backend/..."
	@$(MAKE) -C backend build-docker
	@echo "Running build-docker in frontend/..."
	@$(MAKE) -C frontend build-docker
	@echo "Bringing up containers..."
	@docker compose up -d

down:
	@echo "Stopping containers..."
	@docker compose down

build:
	@echo "Running build-docker in backend/..."
	@$(MAKE) -C backend build-docker
	@echo "Running build-docker in frontend/..."
	@$(MAKE) -C frontend build-docker

clean:
	@echo "Stopping containers and removing volumes..."
	@docker compose down --volumes

