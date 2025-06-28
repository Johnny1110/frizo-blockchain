# Makefile for frizo-blockchain

# Project variables
PROJECT_NAME := frizo-blockchain
BINARY_NAME := frizo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go variables
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
BUILD_DIR := build
DIST_DIR := dist
DOCKER_DIR := docker

# Docker variables
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)
DOCKER_REGISTRY := # Set this if you want to push to a registry

# OS and Architecture for cross-compilation
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: help
help: ## Show this help message
	@echo "$(PROJECT_NAME) Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	$(GO) clean -cache
	docker system prune -f 2>/dev/null || true

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		$(GO) vet ./...; \
	fi

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -cover ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	$(GO) test $(GOFLAGS) -race -cover -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	mkdir -p $(BUILD_DIR)
	$(GO) test -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated: $(BUILD_DIR)/coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

.PHONY: build
build: clean deps ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

.PHONY: build-all
build-all: clean deps ## Build binaries for all platforms
	@echo "Building for all platforms..."
	mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	
	# macOS ARM64 (Apple Silicon)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	
	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) ./cmd/$(BINARY_NAME)

.PHONY: run
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: release
release: test lint build-all ## Create a release (run tests, lint, and build for all platforms)
	@echo "Creating release $(VERSION)..."
	mkdir -p $(DIST_DIR)
	
	# Create archives for each platform
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip -q ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	
	# Generate checksums
	cd $(DIST_DIR) && sha256sum * > $(BINARY_NAME)-$(VERSION)-checksums.txt
	
	@echo "Release artifacts created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

.PHONY: docker-build-multistage
docker-build-multistage: ## Build Docker image using multi-stage build
	@echo "Building Docker image with multi-stage build..."
	docker build -f Dockerfile.multistage -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "DOCKER_REGISTRY not set. Please set it to push images."; \
		exit 1; \
	fi
	@echo "Pushing Docker image to $(DOCKER_REGISTRY)..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest 2>/dev/null || true
	docker system prune -f

.PHONY: dev
dev: fmt lint test build ## Run development workflow (format, lint, test, build)

.PHONY: ci
ci: deps fmt lint test-coverage build ## Run CI workflow

.PHONY: security
security: ## Run security checks
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

.PHONY: mod-update
mod-update: ## Update all Go modules
	@echo "Updating Go modules..."
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: version
version: ## Show version information
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell $(GO) version)"

# Default target
.DEFAULT_GOAL := help