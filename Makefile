# Variables
BINARY_NAME=tipsy
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) .

# Build with version info
.PHONY: build-versioned
build-versioned:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download

# Install binary to GOPATH/bin
.PHONY: install
install:
	$(GOCMD) install .

# Run the application
.PHONY: run
run:
	$(GOCMD) run . $(ARGS)

# Cross-platform builds
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Development workflow
.PHONY: dev
dev: fmt tidy test build

# CI workflow
.PHONY: ci
ci: fmt tidy test lint build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-versioned- Build with version information"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  tidy           - Tidy dependencies"
	@echo "  deps           - Download dependencies"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Run the application (use ARGS=... for arguments)"
	@echo "  build-all      - Build for all platforms"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-darwin   - Build for macOS"
	@echo "  build-windows  - Build for Windows"
	@echo "  dev            - Development workflow (fmt, tidy, test, build)"
	@echo "  ci             - CI workflow (fmt, tidy, test, lint, build)"
	@echo "  help           - Show this help message" 