.PHONY: build clean test install run snapshot release help

# Variables
BINARY_NAME=istfproj
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.DEFAULT_GOAL := help

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "Install complete"

## run: Build and run the binary
run: build
	./$(BINARY_NAME) .

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running golangci-lint..."
	golangci-lint run
	@echo "Lint complete"

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	go mod tidy
	@echo "Tidy complete"

## snapshot: Build snapshot release with GoReleaser
snapshot:
	@echo "Building snapshot release..."
	goreleaser build --snapshot --clean --single-target
	@echo "Snapshot build complete"

## snapshot-all: Build snapshot release for all platforms
snapshot-all:
	@echo "Building snapshot release for all platforms..."
	goreleaser build --snapshot --clean
	@echo "Snapshot build complete"

## release-dry: Dry run of GoReleaser release
release-dry:
	@echo "Running GoReleaser dry run..."
	goreleaser release --snapshot --clean
	@echo "Dry run complete"

## release: Create a new release (requires tag)
release:
	@echo "Creating release..."
	goreleaser release --clean
	@echo "Release complete"

## check: Check GoReleaser configuration
check:
	@echo "Checking GoReleaser configuration..."
	goreleaser check
	@echo "Check complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

## all: Run fmt, vet, test, and build
all: fmt vet test build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
