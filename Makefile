# Suppress Checker Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X suppress-checker/pkg/version.Version=$(VERSION) -X suppress-checker/pkg/version.GitCommit=$(COMMIT) -X suppress-checker/pkg/version.BuildDate=$(BUILD_DATE)"

# Build settings
BINARY_NAME := suppress-checker
BUILD_DIR := build
MAIN_FILE := main.go

# Go settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all build clean test lint fmt vet help install deps release

## Default target
all: build

## Build the application
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

## Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "✅ Built all platform binaries in $(BUILD_DIR)/"

## Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

## Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## Lint the code
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		@echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

## Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ Vet completed"

## Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "✅ Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin"

## Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

## Run the application locally
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(LDFLAGS) $(MAIN_FILE) $(ARGS)

## Run with example
run-example:
	@echo "Running with example directory..."
	go run $(LDFLAGS) $(MAIN_FILE) check --dir examples --verbose

## Show version information
version:
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) version; \
	else \
		echo "Binary not found. Run 'make build' first."; \
	fi

## Create a release (requires git tag)
release: clean build-all
	@echo "Creating release..."
	@if [ -z "$(shell git tag --points-at HEAD)" ]; then \
		echo "❌ No git tag found at HEAD. Create a tag first: git tag v0.1.0"; \
		exit 1; \
	fi
	@echo "✅ Release built for tag: $(shell git tag --points-at HEAD)"

## Show help
help:
	@echo "Suppress Checker - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build the application"
	@echo "  make run-example              # Run with example files"
	@echo "  make test                     # Run tests"
	@echo "  make ARGS='--help' run        # Run with custom arguments"

## Show current version
show-version:
	@echo "Current version: $(VERSION)"
	@echo "Git commit: $(COMMIT)"
	@echo "Build date: $(BUILD_DATE)" 