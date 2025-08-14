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
- `tasks-json-cli validate [path]` - Validate tasks.json syntax and structure
- `tasks-json-cli watch <task-name>` - Watch files and auto-execute task

### Global Flags
- `--config, -c` - Specify tasks.json file path
- `--verbose, -v` - Verbose output
- `--quiet, -q` - Minimal output

## Validation Features

The `validate` command provides comprehensive validation for tasks.json configuration files:

### Validation Types

**Errors** (cause validation failure with exit code 1):
- File not found
- Invalid JSON syntax
- Missing required fields (label, type, command)
- Duplicate task labels

**Warnings** (validation passes but issues are reported):
- Unknown task types (not 'shell' or 'process')
- References to non-existent dependency tasks
- Non-existent working directories (absolute paths only)

### Usage Examples
```bash
# Validate current directory's tasks.json
tasks-json-cli validate

# Validate specific file
tasks-json-cli validate path/to/tasks.json

# Quiet mode (exit code only)
tasks-json-cli validate --quiet

# Verbose output with additional details
tasks-json-cli validate --verbose
```

## Watch Features

The `watch` command provides file monitoring capabilities that automatically execute tasks when file changes are detected:

### Watch Options

**File Monitoring**:
- Real-time file change detection using fsnotify
- Configurable watch paths (defaults to workspace folder)
- File extension filtering (e.g., `.go`, `.js`, `.ts`)
- Path exclusion patterns (defaults: `node_modules`, `.git`, `.vscode`)

**Execution Control**:
- Debounced execution with configurable delay (default: 500ms)
- Support for all VS Code variable substitution
- Verbose and quiet output modes

### Usage Examples
```bash
# Watch workspace folder for any changes and run build task
tasks-json-cli watch build

# Watch specific paths with file extension filtering
tasks-json-cli watch test --path src --path tests --ext .go,.js

# Watch with custom delay and exclusions
tasks-json-cli watch compile --delay 1s --exclude node_modules,dist,build

# Watch with workspace folder override
tasks-json-cli watch lint --workspace-folder /path/to/project

# Verbose watch mode
tasks-json-cli watch format --verbose
```

### Watch Command Flags
- `--path strings`: Paths to watch (defaults to workspace folder)
- `--ext strings`: File extensions to watch (e.g., `.go,.js`)
- `--exclude strings`: Paths to exclude from watching (default: `[node_modules,.git,.vscode]`)
- `--delay duration`: Delay before executing task after file change (default: `500ms`)
- `--workspace-folder string`: Workspace folder path (defaults to git root)
- `--file string`: File path to replace `${file}` variable

## Dependency Support

The CLI now supports VS Code's compound task dependencies through the `dependsOn` property, enabling automatic execution of prerequisite tasks.

### Dependency Features

**Task Dependencies**:
- Single dependency: `"dependsOn": "task-name"`
- Multiple dependencies: `"dependsOn": ["task1", "task2"]`
- Execution order control: `"dependsOrder": "sequence" | "parallel"` (default: parallel)

**Dependency Resolution**:
- Topological sort algorithm ensures correct execution order
- Circular dependency detection with detailed error messages
- Missing dependency validation and reporting
- Support for complex dependency graphs

**Enhanced Commands**:
- `run` command automatically resolves and executes dependencies
- `--dry-run` shows complete execution plan with dependency order
- `validate` command checks for dependency issues

### Dependency Examples
```bash
# Run task with dependencies - automatically resolves and executes prerequisites
tasks-json-cli run deploy

# Show execution plan including all dependencies
tasks-json-cli run test --dry-run

# Validate dependency configuration
tasks-json-cli validate
```

### Dependency Configuration Example
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "clean",
      "type": "shell",
      "command": "rm -rf build/"
    },
    {
      "label": "compile", 
      "type": "shell",
      "command": "go build -o build/app",
      "dependsOn": "clean"
    },
    {
      "label": "test",
      "type": "shell",
      "command": "go test ./...",
      "dependsOn": ["compile", "lint"],
      "dependsOrder": "parallel"
    },
    {
      "label": "deploy",
      "type": "shell",
      "command": "kubectl apply -f deployment.yaml",
      "dependsOn": "test"
    }
  ]
}
```

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
- `init` command with template support ✓
- `info` command for task details ✓
- `validate` command for tasks.json validation ✓

### Phase 3: Advanced Features
- `watch` command with file monitoring ✓
- Compound task (dependsOn) support ✓
- npm/typescript task type support
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