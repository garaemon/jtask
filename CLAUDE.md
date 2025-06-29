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

jtask is a CLI tool written in Go that executes tasks defined in VS Code's `tasks.json` configuration files.

### Current Structure
- `main.go`: Entry point (currently just a placeholder)
- Go module: `github.com/garaemon/jtask`
- Target Go version: 1.24.4
- CI/CD: GitHub Actions with build and lint jobs

### Planned Directory Structure
```
jtask/
├── main.go              # Entry point
├── cmd/                 # Command definitions (cobra)
│   ├── root.go          # Root command
│   ├── list.go          # list command
│   ├── run.go           # run command
│   ├── init.go          # init command
│   ├── info.go          # info command
│   ├── validate.go      # validate command
│   └── watch.go         # watch command
├── internal/            # Internal packages
│   ├── config/          # Configuration file handling
│   │   ├── parser.go    # tasks.json parser
│   │   └── types.go     # Task type definitions
│   ├── executor/        # Task execution engine
│   │   ├── shell.go     # Shell task execution
│   │   ├── process.go   # Process task execution
│   │   └── runner.go    # Execution coordinator
│   └── discovery/       # Task file discovery
│       └── finder.go    # .vscode/tasks.json search
└── templates/           # init templates
    ├── default.json
    ├── go.json
    └── node.json
```

## CLI Command Structure

### Core Commands
- `jtask list [--group <group>] [--type <type>]` - List available tasks
- `jtask run <task-name> [--dry-run]` - Execute specified task
- `jtask init [--template <template>]` - Initialize basic tasks.json file
- `jtask info <task-name>` - Show task details
- `jtask validate [path]` - Validate tasks.json syntax
- `jtask watch <task-name>` - Watch files and auto-execute task

### Global Flags
- `--config, -c` - Specify tasks.json file path
- `--verbose, -v` - Verbose output
- `--quiet, -q` - Minimal output

## Implementation Phases

### Phase 1: Basic Functionality
- tasks.json parser
- Task discovery functionality
- `list` command
- `run` command (shell/process tasks)

### Phase 2: Extended Features
- `init` command
- `info` command
- `validate` command

### Phase 3: Advanced Features
- `watch` command
- npm/typescript task type support
- Compound task (dependsOn) support

## Key Design Goals

The tool aims to bridge VS Code task configurations with command-line execution, allowing developers to run VS Code tasks without the editor.

## Development Notes

- Uses Cobra framework for CLI structure
- Auto-discovers .vscode/tasks.json files
- Extensible execution engine for different task types
- Template-based init functionality
- The project uses golangci-lint for code quality
- CI runs on both build and lint jobs
- Currently uses Go 1.21 in CI but go.mod specifies 1.24.4