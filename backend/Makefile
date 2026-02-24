.PHONY: help generate build-docker remove-docker run

# Display this help message
help:
	@echo "Available targets:"
	@echo "  make generate       - Generate the go code for the database and API"
	@echo "  make build-docker   - Build the Docker image for the API server locally"
	@echo "  make remove-docker  - Remove the local Docker image for the API server"
	@echo "  make run            - Run the API server locally directly through go"
	@echo "  make help           - Display this help message"

# Generate the go code for the database and API
generate:
	go tool sqlc -f db/sqlc.yaml generate
	go tool oapi-codegen -config api/oapi-codegen.yaml ../swagger/api.yaml

# Build the Docker image for the API server locally
build-docker:
	docker build -t alexsieland/bg-library:dev .

# Remove the local Docker image for the API server locally
remove-docker:
	docker rmi alexsieland/bg-library:dev

# Run the API server locally directly through go (without Docker)
run:
	go run main.go
