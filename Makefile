# Build variables
BINARY_NAME=libgo-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=bin
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}"

# Go commands
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOVET=$(GO) vet
GOGET=$(GO) get
GOFMT=$(GO) fmt
GOIMPORTS=goimports

# Test variables
TEST_FLAGS=-v -race
COVERAGE_PROFILE=coverage.out
COVERAGE_HTML=coverage.html

# Tool versions (check these against requirements)
GO_VERSION=1.24.0
GOLANGCI_LINT_VERSION=1.56.2
GOSEC_VERSION=2.19.0
GOVULNCHECK_VERSION=1.0.2
MOCKGEN_VERSION=0.4.0

# Linter configuration
GOLANGCI_LINT=golangci-lint
GOSEC=gosec
GOVULNCHECK=govulncheck
MOCKGEN=mockgen

# Docker configuration
DOCKER=docker
DOCKER_IMAGE=libgo-server
DOCKER_TAG=$(VERSION)

# Source directories
SRC_DIRS=./cmd/... ./internal/... ./pkg/...

# Interfaces to generate mocks for
MOCK_INTERFACES=internal/libvirt/connection/interface.go internal/libvirt/domain/interface.go \
                internal/libvirt/storage/interface.go internal/libvirt/network/interface.go \
                internal/vm/interface.go internal/vm/template/interface.go \
                internal/vm/cloudinit/interface.go internal/export/interface.go \
                internal/auth/jwt/claims.go internal/auth/user/service_interface.go \
                pkg/logger/interface.go

.PHONY: all build clean test unit-test integration-test coverage lint sec-scan vuln-check mocks help docker-build docker-run setup

all: test build

setup: install-tools ## Set up development environment

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) golang.org/x/tools/cmd/goimports
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	$(GOGET) github.com/securego/gosec/v2/cmd/gosec@v$(GOSEC_VERSION)
	$(GOGET) golang.org/x/vuln/cmd/govulncheck@v$(GOVULNCHECK_VERSION)
	$(GOGET) go.uber.org/mock/mockgen@v$(MOCKGEN_VERSION)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

build-dev: ## Build with debugging information
	@echo "Building $(BINARY_NAME) (development)..."
	$(GOBUILD) $(LDFLAGS) -race -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_PROFILE) $(COVERAGE_HTML)

version-check: ## Verify required tool versions
	@echo "Checking Go version..."
	@go version | grep -q "go$(GO_VERSION)" || (echo "Error: Required Go version is $(GO_VERSION)" && exit 1)
	@echo "Go version check passed"
	@echo "Checking golangci-lint version..."
	@golangci-lint --version | grep -q "$(GOLANGCI_LINT_VERSION)" || (echo "Warning: Recommended golangci-lint version is $(GOLANGCI_LINT_VERSION)" && exit 0)
	@echo "golangci-lint version check passed"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	$(GOIMPORTS) -w ./cmd ./internal ./pkg ./test

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

lint: ## Run linters
	@echo "Running linters..."
	$(GOLANGCI_LINT) run ./...

sec-scan: ## Run security scan
	@echo "Running security scan..."
	$(GOSEC) ./...

vuln-check: ## Run vulnerability check
	@echo "Running vulnerability check..."
	$(GOVULNCHECK) ./...

test: ## Run all tests with coverage
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) -coverprofile=$(COVERAGE_PROFILE) ./...

unit-test: ## Run only unit tests
	@echo "Running unit tests..."
	$(GOTEST) -short $(TEST_FLAGS) ./...

integration-test: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -run Integration $(TEST_FLAGS) ./test/integration/...

test-ubuntu-docker: ## Run the Ubuntu Docker deployment test
	@chmod +x ./test/integration/run-docker-test.sh
	@./test/integration/run-docker-test.sh

coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$(GO) tool cover -func=$(COVERAGE_PROFILE)

mocks: install-tools ## Generate mock implementations for interfaces
	@echo "Generating mocks..."
	@for interface in $(MOCK_INTERFACES); do \
		PACKAGE=$$(echo $$interface | cut -d'/' -f2) && \
		GOFILE=$$(basename $$interface) && \
		echo "Generating mock for $$interface..." && \
		$(MOCKGEN) -source=$$interface -destination=./test/mocks/$$PACKAGE/$$GOFILE -package=mocks_$$PACKAGE; \
	done

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	$(DOCKER) run -p 8080:8080 --name $(BINARY_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean: ## Clean Docker artifacts
	@echo "Cleaning Docker artifacts..."
	$(DOCKER) rm -f $(BINARY_NAME) 2>/dev/null || true
	$(DOCKER) rmi -f $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
