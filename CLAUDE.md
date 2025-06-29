# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building
```bash
go build -o jtask        # Build the executable
go install              # Build and install
```

### Testing
```bash
go test -v ./...         # Run all tests with verbose output
```

### Linting
```bash
golangci-lint run        # Run linter (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
```

## Project Architecture

jtask is a CLI tool written in Go that executes tasks defined in VS Code's `tasks.json` configuration files. The project is currently in early development with a minimal structure:

- `main.go`: Entry point (currently just a placeholder)
- Go module: `github.com/garaemon/jtask`
- Target Go version: 1.24.4
- CI/CD: GitHub Actions with build and lint jobs

## Key Design Goals

The tool aims to bridge VS Code task configurations with command-line execution, allowing developers to run VS Code tasks without the editor.

## Development Notes

- The project uses golangci-lint for code quality
- CI runs on both build and lint jobs
- Currently uses Go 1.21 in CI but go.mod specifies 1.24.4