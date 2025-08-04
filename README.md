# Lit

A simple CLI agent powered by Claude that can interact with your filesystem and search through code.

## Features

- **File Operations**: Read, list, edit, move, and remove files with safety checks
- **Code Search**: Search patterns across files using ripgrep
- **File Discovery**: Fast file finding by name using fd
- **Git Integration**: Status, add, commit, and diff operations
- **Interactive Chat**: Natural language interface with Claude
- **Tool Safety**: Prevents accidental overwrites and requires confirmation for destructive operations

## Installation

### Quick Install (Linux/macOS)
```bash
curl -sSf https://raw.githubusercontent.com/carlosarraes/lit/main/install.sh | sh
```

### Manual Build
```bash
make build
```

This builds the binary and copies it to `~/.local/bin/lit`.

### Manual Download
Download the binary for your platform from the [releases page](https://github.com/carlosarraes/lit/releases).

## Usage

```bash
lit
```

Start chatting with Claude. Available commands:
- Ask Claude to read files: "Show me the contents of main.go"
- Search for patterns: "Find all TODO comments in the codebase"
- Find files by name: "Find all test files in the project"
- Edit files: "Add a comment to the top of this file"
- Move/rename files: "Rename config.json to settings.json"
- Remove files: "Delete the temp directory"
- Git operations: "Check git status and commit changes"
- List directories: "What files are in the src folder?"

## Requirements

- Go 1.24.5+
- Anthropic API key (set via environment variable)
- ripgrep (`rg`) for search functionality
- fd for fast file finding

## Tools

- `read_file` - Read file contents with safety checks
- `list_files` - List directory contents
- `edit_file` - Edit files with read-before-edit validation
- `ripgrep` - Search patterns across files using ripgrep
- `fd` - Find files and directories by name using fd
- `rm` - Remove files and directories with user confirmation
- `mv` - Move and rename files and directories
- `git_status` - Show git working tree status
- `git_add` - Add files to git staging area
- `git_commit` - Create git commits with messages
- `git_diff` - Show git differences between versions
