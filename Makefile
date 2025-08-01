# Lit - Claude CLI Agent
.PHONY: help build install clean test fmt check

BINARY_NAME=lit
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION)"

help: ## Show this help message
	@echo "Lit - Claude CLI Agent"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary and copy to ~/.local/bin
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/lit
	@mkdir -p ~/.local/bin
	@cp $(BINARY_NAME) ~/.local/bin/
	@echo "Copied to ~/.local/bin/$(BINARY_NAME)"

install: build ## Build and install to GOPATH/bin
	go install ./cmd/lit

test: ## Run tests
	go test -v ./...

fmt: ## Format code
	go fmt ./...

check: fmt test ## Run all checks (format, test)

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
