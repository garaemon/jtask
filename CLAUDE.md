# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building
```bash
go build -o tasks-json-cli        # Build the executable
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

tasks-json-cli is a CLI tool written in Go that executes tasks defined in VS Code's `tasks.json` configuration files.

### Current Structure
- `main.go`: Entry point (currently just a placeholder)
- Go module: `github.com/garaemon/tasks-json-cli`
- Target Go version: 1.24.4
- CI/CD: GitHub Actions with build and lint jobs

### Planned Directory Structure
```
tasks-json-cli/
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
│   ├── variables/       # Variable resolution system
│   │   ├── resolver.go  # Main variable resolver
│   │   ├── context.go   # Variable context holder
│   │   ├── file.go      # File-related variables
│   │   ├── workspace.go # Workspace variables
│   │   ├── environment.go # Environment variables
│   │   └── system.go    # System variables
│   └── discovery/       # Task file discovery
│       └── finder.go    # .vscode/tasks.json search
└── templates/           # init templates
    ├── default.json
    ├── go.json
    └── node.json
```

## CLI Command Structure

### Core Commands
- `tasks-json-cli list [--group <group>] [--type <type>]` - List available tasks
- `tasks-json-cli run <task-name> [--dry-run]` - Execute specified task
- `tasks-json-cli init [--template <template>]` - Initialize basic tasks.json file
- `tasks-json-cli info <task-name>` - Show task details
- `tasks-json-cli validate [path]` - Validate tasks.json syntax
- `tasks-json-cli watch <task-name>` - Watch files and auto-execute task

### Global Flags
- `--config, -c` - Specify tasks.json file path
- `--verbose, -v` - Verbose output
- `--quiet, -q` - Minimal output

## Variable Support

VS Code tasks.json supports extensive variable substitution. tasks-json-cli now supports all major VS Code file-related variables.

### Supported Variable Support
- `${workspaceFolder}` - Path to workspace folder ✓
- `${file}` - Path to currently selected file (via --file flag) ✓
- `${cwd}` - Current working directory ✓
- `${pathSeparator}` - OS-specific path separator ✓
- `${env:VARNAME}` - Environment variable expansion ✓
- `${workspaceFolderBasename}` - Workspace folder name only ✓
- `${fileBasename}` - Current file name with extension ✓
- `${fileBasenameNoExtension}` - Current file name without extension ✓
- `${fileDirname}` - Directory path of current file ✓
- `${fileExtname}` - Extension of current file ✓
- `${fileWorkspaceFolder}` - Workspace folder of the current file ✓
- `${relativeFile}` - Current file relative to workspace root ✓
- `${relativeFileDirname}` - Directory of current file relative to workspace ✓

### VS Code Variables Not Applicable to CLI Context
- `${execPath}` - Path to VS Code executable (not applicable)
- `${config:setting}` - VS Code configuration values (not applicable)
- `${lineNumber}` - Current line number in editor (not applicable)
- `${selectedText}` - Currently selected text (not applicable)

### Proposed Variable Architecture

#### New Package Structure
```
internal/
└── variables/           # Variable resolution system
    ├── resolver.go      # Main variable resolver
    ├── context.go       # Variable context holder
    ├── file.go          # File-related variables
    ├── workspace.go     # Workspace variables
    ├── environment.go   # Environment variables
    └── system.go        # System variables
```

#### Core Components
- `VariableContext` - Holds workspace, file, and environment state
- `VariableResolver` interface - Pluggable resolver pattern
- `ResolveAllVariables(text, context)` - Main resolution function

#### Integration Points
- Replace `substituteVariables()` in `internal/executor/runner.go`
- Update both run command and dry-run functionality
- Maintain backward compatibility with existing variables

## Implementation Phases

### Phase 1: Core Functionality (COMPLETED)
- tasks.json parser ✓
- Task discovery functionality ✓
- `list` command ✓
- `run` command (shell/process tasks) ✓

### Phase 1.5: Enhanced Variable Support (COMPLETED)
**This phase should be completed before adding new commands**
- ✓ Support all VS Code file-related variables
- ✓ Update existing `run` command to use new variable system
- ✓ Ensure backward compatibility with current variables
- ✓ Added `${cwd}` variable support
- ✓ Added `${pathSeparator}` variable support
- ✓ Added `${env:VARNAME}` environment variable expansion
- ✓ Added `${workspaceFolderBasename}` variable support
- ✓ Added `${fileBasename}` variable support
- ✓ Added `${fileBasenameNoExtension}` variable support
- ✓ Added `${fileDirname}` variable support
- ✓ Added `${fileExtname}` variable support
- ✓ Added `${fileWorkspaceFolder}` variable support
- ✓ Added `${relativeFile}` variable support
- ✓ Added `${relativeFileDirname}` variable support

**Future Enhancement:**
- Implement comprehensive variable resolution system
- Add `internal/variables` package with pluggable resolvers

### Phase 2: Extended Commands
- `init` command with template support
- `info` command for task details
- `validate` command for tasks.json validation

### Phase 3: Advanced Features
- `watch` command with file monitoring
- npm/typescript task type support
- Compound task (dependsOn) support
- Task groups and organization features

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