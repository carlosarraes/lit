.PHONY: build 

BINARY_NAME=lit
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION)"

build: ## Build the binary and copy to ~/.local/bin
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/lit
	@mkdir -p ~/.local/bin
	@cp $(BINARY_NAME) ~/.local/bin/
	@echo "Copied to ~/.local/bin/$(BINARY_NAME)"
