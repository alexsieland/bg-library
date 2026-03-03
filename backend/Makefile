.PHONY: help generate build-docker remove-docker run format lint test

# Display this help message
help:
	@echo "Available targets:"
	@echo "  make generate       - Generate the go code for the database and API"
	@echo "  make build-docker   - Build the Docker image for the API server locally"
	@echo "  make remove-docker  - Remove the local Docker image for the API server"
	@echo "  make run            - Run the API server locally directly through go"
	@echo "  make format         - Format and simplify code using gofmt and goimports"
	@echo "  make lint           - Check code formatting"
	@echo "  make test           - Run all unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make help           - Display this help message"

# Generate the go code for the database and API
generate:
	@echo "Generating go code for database and API..."
	go tool sqlc -f db/sqlc.yaml generate
	go tool oapi-codegen -config api/oapi-codegen.yaml ../swagger/api.yaml

# Build the Docker image for the API server locally
build-docker:
	@echo "Building local Docker image for API server..."
	cd .. && docker build -f backend/Dockerfile -t ghcr.io/${GHCR_USERNAME:-alexsieland}/bg-library-backend:dev .

# Remove the local Docker image for the API server locally
remove-docker:
	@echo "Removing local Docker image for API server..."
	docker rmi ghcr.io/${GHCR_USERNAME:-alexsieland}/bg-library-backend:dev

# Run the API server locally directly through go (without Docker)
run:
	@echo "Running API server locally..."
	go run main.go

# Format and simplify code
format:
	@echo "Formatting code..."
	@gofmt -s -w .
	@go run golang.org/x/tools/cmd/goimports@latest -w .

# Check code formatting
lint:
	@echo "Checking code formatting..."
	@gofmt -s -l . | grep -v "gen.go" || true

# Run unit tests (use TEST_ARGS="-run TestName" to run specific tests)
test:
	@echo "Running unit tests..."
	@go test -v ./... $(TEST_ARGS)

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/... $(TEST_ARGS)
