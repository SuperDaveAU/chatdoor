.PHONY: build server client clean help

BINARY_NAME=chatdoor

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) ./cmd/chatdoor

server: build ## Run in server mode
	@./bin/$(BINARY_NAME) server

client: build ## Run in client mode (requires -addr and -pubkey)
	@./bin/$(BINARY_NAME) client

test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f chatdoor_private.key chatdoor_public.key

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

install: build ## Install binary to $GOPATH/bin
	@echo "Installing..."
	go install ./cmd/chatdoor

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run
