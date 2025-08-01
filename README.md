# Lit

A simple CLI agent powered by Claude that can interact with your filesystem and search through code.

## Features

- **File Operations**: Read, list, and edit files
- **Code Search**: Search patterns across files using ripgrep
- **Interactive Chat**: Natural language interface with Claude
- **Tool Safety**: Prevents accidental file overwrites by requiring read-before-edit

## Installation

```bash
make build
```

This builds the binary and copies it to `~/.local/bin/lit`.

## Usage

```bash
lit
```

Start chatting with Claude. Available commands:
- Ask Claude to read files: "Show me the contents of main.go"
- Search for patterns: "Find all TODO comments in the codebase"
- Edit files: "Add a comment to the top of this file"
- List directories: "What files are in the src folder?"

## Requirements

- Go 1.24.5+
- Anthropic API key (set via environment)
- ripgrep (`rg`) for search functionality

## Tools

- `read_file` - Read file contents
- `list_files` - List directory contents  
- `edit_file` - Edit files with validation
- `ripgrep` - Search patterns in files
