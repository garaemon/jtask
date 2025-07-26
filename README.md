# tasks-json-cli

A CLI tool written in Go to execute tasks defined in `tasks.json` configuration files (VS Code task format).

## Purpose

tasks-json-cli allows you to run tasks from VS Code's `tasks.json` configuration directly from the command line, providing a convenient way to execute build scripts, tests, and other development tasks without needing VS Code.

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