# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Installation
- `make build` - Build the binary and copy to ~/.local/bin/lit
- `make install` - Build and install to GOPATH/bin  
- `make clean` - Clean build artifacts

### Testing and Quality
- `make test` - Run tests with `go test -v ./...`
- `make fmt` - Format code with `go fmt ./...`
- `make check` - Run all checks (format + test)

### Running the Application
- `lit` - Start the interactive CLI agent (requires Anthropic API key in environment)

## Architecture Overview

This is a Go CLI application that provides an interactive chat interface with Claude AI, enhanced with filesystem tools.

### Core Components

**Main Entry Point** (`cmd/lit/main.go`):
- Initializes Anthropic client and tools
- Sets up the agent with predefined tool definitions
- Handles the main conversation loop

**Agent System** (`internal/agent/agent.go`):
- Manages conversation flow with Claude
- Handles tool execution and results
- Supports interactive input with file completion
- Processes @ references for file paths
- Uses Claude 3.5 Haiku model

**Tool System** (`internal/tools/`):
- `readFile.go` - Read file contents with safety checks
- `listFile.go` - List directory contents
- `editFile.go` - Edit files with read-before-edit validation
- `ripgrep.go` - Search patterns across files using ripgrep
- `fd.go` - Find files and directories by name using fd
- `rm.go` - Remove files and directories with user confirmation
- `mv.go` - Move and rename files and directories
- `git.go` - Git operations (status, add, commit, diff)
- `defitinions.go` - Tool schema generation using JSON Schema

**Interactive Input** (`internal/input/`):
- Terminal-based input with file path completion
- Supports @ references for file selection
- Multi-line input support with continuation
- Custom readline implementation using golang.org/x/term

### Key Design Patterns

**Tool Safety**: The edit tool requires reading a file before editing to prevent accidental overwrites. The rm tool requires explicit user confirmation (y/N) before deleting files.

**@ Reference Processing**: User input is processed to convert @filename patterns into file paths for easy file referencing.

**Schema Generation**: Tools use reflection and JSON Schema to automatically generate parameter schemas for Claude.

## Dependencies

- `github.com/anthropics/anthropic-sdk-go` - Official Anthropic SDK
- `github.com/invopop/jsonschema` - JSON Schema generation
- `golang.org/x/term` - Terminal control for interactive input
- External dependency: `ripgrep` command-line tool for search functionality
- External dependency: `fd` command-line tool for fast file finding

## Environment Requirements

- Go 1.24.5+
- Anthropic API key (set via environment variable)
- ripgrep (`rg`) installed for search functionality
- fd installed for fast file finding