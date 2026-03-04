PHONY: help generate build-docker remove-docker run format lint test

export GHCR_USERNAME ?= alexsieland

help:
	@echo "Available commands:"
	@echo "  generate       - Generate TypeScript types from OpenAPI spec"
	@echo "  build-docker   - Build the local Docker image for the API server"
	@echo "  remove-docker  - Remove the local Docker image for the API server"
	@echo "  run            - Run the frontend locally"
	@echo "  format         - Format the code using Prettier"
	@echo "  lint           - Check code formatting using Prettier"
	@echo "  test           - Run tests"

generate:
	@echo "Generating code for frontend..."
	@npx openapi-typescript ../swagger/api.yaml --output src/generated/library-api.d.ts

build-docker:
	@echo "Building local Docker image for API server..."
	cd .. && docker build -f frontend/Dockerfile -t ghcr.io/$(GHCR_USERNAME)/bg-library-frontend:dev .

remove-docker:
	@echo "Removing local Docker image for API server..."
	docker rmi ghcr.io/$(GHCR_USERNAME)/bg-library-frontend:dev

run:
	@echo "Running frontend locally..."
	npm run dev

format:
	@echo "Formatting code..."
	@npx prettier --write .

lint:
	@echo "Checking code formatting..."
	@npx prettier --check .

test:
	@echo "Running tests..."
	@npx run
