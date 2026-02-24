.PHONY: up down down-clean help

help:
	@echo "Available targets:"
	@echo "  make up          - Build and start the containers"
	@echo "  make down        - Stop the containers"
	@echo "  make clean       - Clean docker volumes (will stop containers if they are running)"
	@echo "  make help        - display this message"

up:
	@echo "Running generate and build-docker in backend/..."
	@$(MAKE) -C backend generate build-docker
	@echo "Bringing up containers..."
	@docker compose up -d

down:
	@echo "Stopping containers..."
	@docker compose down

clean:
	@echo "Stopping containers and removing volumes..."
	@docker compose down --volumes
