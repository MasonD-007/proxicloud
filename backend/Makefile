# ProxiCloud Backend Makefile

.PHONY: help dev build test clean lint fmt install deps run

# Variables
BINARY_NAME=api
MAIN_PATH=cmd/api/main.go
BUILD_DIR=bin
DIST_DIR=../dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

install: deps ## Install dependencies
	@echo "Dependencies installed"

dev: ## Run in development mode with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build for Linux (amd64)
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/proxicloud-api-linux-amd64 $(MAIN_PATH)
	@echo "Binary built: $(DIST_DIR)/proxicloud-api-linux-amd64"

build-linux-arm: ## Build for Linux (arm64)
	@echo "Building for Linux (arm64)..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/proxicloud-api-linux-arm64 $(MAIN_PATH)
	@echo "Binary built: $(DIST_DIR)/proxicloud-api-linux-arm64"

build-all: build-linux build-linux-arm ## Build for all platforms
	@echo "All binaries built successfully"

test: ## Run tests
	$(GOTEST) -v -race ./...

test-cover: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	$(GOTEST) -v -tags=integration ./...

lint: ## Run linter
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	$(GOFMT) -w .

vet: ## Run go vet
	$(GOVET) ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.txt coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	docker build -t proxicloud-api:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env proxicloud-api:latest

watch: ## Watch for changes and rebuild (requires entr)
	@if command -v entr > /dev/null; then \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "entr not installed. Using 'make dev' instead..."; \
		make dev; \
	fi

.DEFAULT_GOAL := help
