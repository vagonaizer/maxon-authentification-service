.PHONY: help build run test clean proto deps docker-build docker-run migrate lint format init reset

APP_NAME := auth-service
VERSION := v1.0.0
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

PROTO_DIR := api/proto
PROTO_OUT_DIR := api/proto/generated
GO_OUT_DIR := $(PROTO_OUT_DIR)

reset: ## Reset and clean everything
	@echo "Resetting project..."
	rm -rf go.sum
	rm -rf bin/
	rm -rf api/proto/generated/*.pb.go
	go clean -modcache
	go mod tidy
	@echo "Project reset completed!"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

init: ## Initialize project (run this first)
	@echo "Initializing project..."
	@if [ ! -d .git ]; then \
		git init; \
		git add .; \
		git commit -m "Initial commit"; \
		echo "Git repository initialized"; \
	fi
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
	fi
	@mkdir -p api/proto/generated
	@mkdir -p bin
	@mkdir -p logs
	@echo "Project initialized successfully!"

deps-only: ## Install only Go dependencies (without proto generation)
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto-clean: ## Clean generated protobuf files
	@echo "Cleaning generated protobuf files..."
	rm -rf $(PROTO_OUT_DIR)/*.pb.go

proto: proto-clean ## Generate protobuf files
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT_DIR)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Protobuf files generated successfully!"

deps: deps-only proto ## Install dependencies and generate proto files
	@echo "Dependencies and proto files ready!"

build: ## Build the application (generates proto if needed)
	@echo "Checking if proto files exist..."
	@if [ ! -f "$(PROTO_OUT_DIR)/auth.pb.go" ] || [ ! -f "$(PROTO_OUT_DIR)/user.pb.go" ]; then \
		echo "Proto files missing, generating..."; \
		$(MAKE) proto; \
	fi
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go
	@echo "Build completed: bin/$(APP_NAME)"

build-migrate: ## Build migration tool
	@echo "Building migration tool..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/migrate cmd/migrate/main.go
	@echo "Migration tool built: bin/migrate"

run: build ## Run the application
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)

run-dev: ## Run the application in development mode
	@echo "Checking if proto files exist..."
	@if [ ! -f "$(PROTO_OUT_DIR)/auth.pb.go" ] || [ ! -f "$(PROTO_OUT_DIR)/user.pb.go" ]; then \
		echo "Proto files missing, generating..."; \
		$(MAKE) proto; \
	fi
	@echo "Running $(APP_NAME) in development mode..."
	go run cmd/server/main.go

run-migrate: build-migrate ## Run database migrations
	@echo "Running database migrations..."
	./bin/migrate

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report: coverage.html"; \
	fi

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -race ./internal/services/... ./pkg/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf $(PROTO_OUT_DIR)/*.pb.go
	go clean -cache
	@echo "Clean completed"

lint: ## Run linter (if available)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

docker-run: ## Run with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker-compose down

dev-setup: init deps-only proto ## Setup development environment
	@echo "Development environment setup completed!"
	@echo "Next steps:"
	@echo "1. Edit .env file with your configuration"
	@echo "2. Run 'make dev-db' to start databases"
	@echo "3. Run 'make migrate-up' to run migrations"
	@echo "4. Run 'make run-dev' to start the service"

dev-db: ## Start development databases
	@echo "Starting development databases..."
	docker-compose up -d postgres redis kafka

migrate-up: build-migrate ## Run database migrations up
	@echo "Running migrations up..."
	./bin/migrate -direction=up

migrate-down: build-migrate ## Run database migrations down
	@echo "Running migrations down..."
	./bin/migrate -direction=down

check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@command -v protoc >/dev/null 2>&1 || { echo "❌ protoc is required but not installed. Please install Protocol Buffers compiler."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ go is required but not installed."; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "❌ docker is required but not installed."; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ docker-compose is required but not installed."; exit 1; }
	@echo "✅ All required tools are installed!"

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

.DEFAULT_GOAL := help
