.PHONY: help run build test lint clean deps docker-build migrate-up migrate-down migrate-create

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=board-api
DOCKER_IMAGE=board-service
DOCKER_TAG=latest
DATABASE_URL ?= postgresql://kanban_service:kanban_pass12345@localhost:5432/wealist_kanban_db

help:  ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

run:  ## Run the application locally (ENV=dev)
	@echo "Running board-service..."
	ENV=dev go run cmd/api/main.go

build:  ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) cmd/api/main.go
	@echo "Build complete: $(BINARY_NAME)"

test:  ## Run tests
	@echo "Running tests..."
	go test -v -race -cover ./...

test-coverage:  ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:  ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

lint-fix:  ## Run golangci-lint with auto-fix
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

deps:  ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

clean:  ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean
	@echo "Clean complete"

docker-build:  ## Build Docker image
	@echo "Building Docker image..."
	docker build -f docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run:  ## Run Docker container locally
	@echo "Running Docker container..."
	docker run --rm -p 8000:8000 \
		-e DATABASE_URL=$(DATABASE_URL) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

migrate-create:  ## Create new migration file (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:  ## Run database migrations up
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@echo "Running migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "Migrations complete"

migrate-down:  ## Rollback migrations (usage: make migrate-down n=1)
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@echo "Rolling back migration(s)..."
	migrate -path migrations -database "$(DATABASE_URL)" down $(if $(n),$(n),1)
	@echo "Rollback complete"

migrate-version:  ## Show current migration version
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@echo "Current migration version:"
	migrate -path migrations -database "$(DATABASE_URL)" version

migrate-force:  ## Force migration version (usage: make migrate-force v=1)
	@if [ -z "$(DATABASE_URL)" ] || [ -z "$(v)" ]; then \
		echo "Usage: make migrate-force v=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(v)..."
	migrate -path migrations -database "$(DATABASE_URL)" force $(v)

fmt:  ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

vet:  ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

mod-verify:  ## Verify dependencies
	@echo "Verifying dependencies..."
	go mod verify
	@echo "Dependencies verified"

all: clean deps build test  ## Clean, download deps, build, and test
	@echo "All tasks complete"
