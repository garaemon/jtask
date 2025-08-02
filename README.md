# tasks-json-cli

A CLI tool written in Go to execute tasks defined in `tasks.json` configuration files (VS Code task format).

## Purpose

tasks-json-cli allows you to run tasks from VS Code's `tasks.json` configuration directly from the command line, providing a convenient way to execute build scripts, tests, and other development tasks without needing VS Code.

## Installation

### Using go install

```bash
go install github.com/garaemon/tasks-json-cli@latest
```

### Building from Source

```bash
git clone https://github.com/garaemon/tasks-json-cli.git
cd tasks-json-cli
go build -o tasks-json-cli
```

## Usage

### Basic Commands

```bash
# List all available tasks
tasks-json-cli list

# Run a specific task
tasks-json-cli run <task-name>

# Show task details
tasks-json-cli info <task-name>

# Validate tasks.json file
tasks-json-cli validate

# Initialize a new tasks.json file
tasks-json-cli init
```

### Options

```bash
# Use a specific tasks.json file
tasks-json-cli list --config path/to/tasks.json

# Verbose output
tasks-json-cli run <task-name> --verbose

# Dry run (show what would be executed)
tasks-json-cli run <task-name> --dry-run
```

## Status

⚠️ **This project is currently under development**

## Development

### Building the Project

To build the tasks-json-cli executable:

```bash
go build -o tasks-json-cli
```

To build and install:

```bash
go install
```

### Running Tests

```bash
go test -v ./...
```

### Running Linter

To run golangci-lint for code quality checks:

```bash
# Install golangci-lint if not already installed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```