.PHONY: help build build-all clean test lint run docker-build docker-up docker-down deps

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_DIR := bin
SERVICES := orchestrator metadata github-discovery document-processor embedding vector-storage notification
GO := go
DOCKER_COMPOSE := docker-compose

# Help target
help: ## Display this help message
	@echo "RepoSync - Go Microservices"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build all services
	@echo "Building all services..."
	@mkdir -p $(BINARY_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GO) build -o $(BINARY_DIR)/$$service ./services/$$service; \
	done
	@echo "Build complete!"

build-linux: ## Build for Linux (cross-compile)
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)/linux
	@for service in $(SERVICES); do \
		echo "Building $$service for Linux..."; \
		GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_DIR)/linux/$$service ./services/$$service; \
	done
	@echo "Linux build complete!"

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -rf data/*.db
	@rm -f coverage.out
	@echo "Clean complete!"

# Test targets
test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint targets
lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete!"

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies downloaded!"

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "Dependencies updated!"

# Run targets
run-orchestrator: ## Run orchestrator service
	@echo "Starting orchestrator..."
	$(GO) run ./services/orchestrator/main.go

run-metadata: ## Run metadata service
	@echo "Starting metadata service..."
	$(GO) run ./services/metadata/main.go

# Docker targets
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build
	@echo "Docker build complete!"

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started! Check health: curl http://localhost:8080/health"

docker-down: ## Stop all services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down
	@echo "Services stopped!"

docker-logs: ## View Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-ps: ## List running containers
	$(DOCKER_COMPOSE) ps

docker-test: ## Run integration tests with Docker Compose
	@echo "Running integration tests..."
	$(DOCKER_COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-runner
	$(DOCKER_COMPOSE) -f docker-compose.test.yml down -v

docker-test-build: ## Build and test Docker images
	@echo "Building test images..."
	$(DOCKER_COMPOSE) -f docker-compose.test.yml build
	@echo "Running tests..."
	$(DOCKER_COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-runner
	$(DOCKER_COMPOSE) -f docker-compose.test.yml down -v

# Setup targets
setup: deps ## Initial setup
	@echo "Setting up RepoSync..."
	@mkdir -p data logs
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from template. Please configure it."; \
	fi
	@echo "Setup complete!"

# Sync targets
sync: ## Trigger manual sync (requires orchestrator running)
	@echo "Triggering sync..."
	@curl -X POST "http://localhost:8080/sync?incremental=true" | jq '.'

sync-full: ## Trigger full sync (no incremental)
	@echo "Triggering full sync..."
	@curl -X POST "http://localhost:8080/sync?incremental=false" | jq '.'

auto-sync: ## Start services with automatic sync (incremental)
	@echo "Starting services with auto-sync (incremental)..."
	@SYNC_INCREMENTAL=true docker-compose up

auto-sync-full: ## Start services with automatic full sync
	@echo "Starting services with auto-sync (full)..."
	@SYNC_INCREMENTAL=false docker-compose up

auto-sync-bg: ## Start services with auto-sync in background
	@echo "Starting services with auto-sync in background..."
	@docker-compose up -d
	@echo "View logs: docker-compose logs -f auto-sync"

health: ## Check health of all services
	@echo "Checking service health..."
	@echo "Orchestrator:"; curl -s http://localhost:8080/health | jq '.'
	@echo "GitHub Discovery:"; curl -s http://localhost:8081/health | jq '.'
	@echo "Document Processor:"; curl -s http://localhost:8082/health | jq '.'
	@echo "Embedding:"; curl -s http://localhost:8083/health | jq '.'
	@echo "Vector Storage:"; curl -s http://localhost:8084/health | jq '.'
	@echo "Notification:"; curl -s http://localhost:8085/health | jq '.'
	@echo "Metadata:"; curl -s http://localhost:8086/health | jq '.'

# Development targets
dev: ## Start development environment
	@echo "Starting development environment..."
	@make docker-up
	@make docker-logs

install: build ## Install binaries to GOPATH
	@echo "Installing binaries..."
	@for service in $(SERVICES); do \
		$(GO) install ./services/$$service; \
	done
	@echo "Installation complete!"
