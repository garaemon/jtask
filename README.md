# jtask

A CLI tool written in Go to execute tasks defined in `tasks.json` configuration files (VS Code task format).

## Purpose

jtask allows you to run tasks from VS Code's `tasks.json` configuration directly from the command line, providing a convenient way to execute build scripts, tests, and other development tasks without needing VS Code.

## Status

⚠️ **This project is currently under development**

## Development

### Building the Project

To build the jtask executable:

```bash
go build -o jtask
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