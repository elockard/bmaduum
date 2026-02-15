# Development Guide

Guide for developing and extending `bmaduum`.

## Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/dl/)
- **just** - Task runner ([Install just](https://github.com/casey/just))
- **golangci-lint** - Linter ([Install golangci-lint](https://golangci-lint.run/))
- **Claude CLI** - For integration testing

### Clone and Build

```bash
git clone https://github.com/ibro45/bmaduum.git
cd bmaduum
go mod download
just build
```

### Verify Setup

```bash
just test
just lint
just check
```

## Project Structure

```
bmaduum/
├── cmd/bmaduum/
│   └── main.go              # Entry point
│
├── internal/
│   ├── cli/                  # CLI commands (Cobra)
│   │   ├── root.go           # Root command, App struct, dependency injection
│   │   ├── story.go          # story command (lifecycle)
│   │   ├── epic.go           # epic command (batch lifecycle)
│   │   ├── workflow_cmd.go   # workflow subcommands (individual steps)
│   │   ├── raw.go            # raw command
│   │   ├── retry.go          # executeWithRetry helper
│   │   ├── errors.go         # ExitError type
│   │   └── test_helpers.go   # Mock types for testing
│   │
│   ├── claude/               # Claude CLI integration
│   │   ├── types.go          # Event types, StreamEvent
│   │   ├── client.go         # Executor interface and DefaultExecutor
│   │   └── parser.go         # JSON stream parser
│   │
│   ├── config/               # Configuration
│   │   ├── types.go          # Config, WorkflowConfig, ClaudeConfig
│   │   └── config.go         # Loader, GetPrompt, GetModel
│   │
│   ├── lifecycle/            # Lifecycle orchestration
│   │   └── executor.go       # Executor with bmad-help fallback
│   │
│   ├── router/               # Workflow routing
│   │   ├── router.go         # Router struct (hardcoded + manifest-driven)
│   │   └── lifecycle.go      # LifecycleStep type
│   │
│   ├── manifest/             # BMAD v6 file discovery
│   │   ├── manifest.go       # Workflow manifest CSV parsing
│   │   └── modules.go        # Module manifest YAML parsing
│   │
│   ├── bmadhelp/             # bmad-help fallback
│   │   └── bmadhelp.go       # ClaudeFallback, ParseResponse
│   │
│   ├── workflow/             # Workflow execution
│   │   ├── workflow.go       # Runner (RunSingle, RunRaw)
│   │   ├── steps.go          # Package doc
│   │   └── tool_correlator.go # Tool use/result correlation
│   │
│   ├── status/               # Sprint status
│   │   ├── types.go          # Status type and constants
│   │   ├── reader.go         # Reader with path discovery
│   │   └── writer.go         # Writer with atomic updates
│   │
│   ├── state/                # State persistence
│   │   └── state.go          # Manager for save/load/clear
│   │
│   ├── ratelimit/            # Rate limit detection
│   │   └── detector.go       # Detector and State
│   │
│   └── output/               # Terminal output
│       ├── core/             # Printer interface, types
│       ├── diff/             # Diff rendering
│       ├── progress/         # Progress line
│       ├── render/           # Specialized renderers
│       └── terminal/         # ANSI control
│
├── config/
│   └── workflows.yaml        # Default configuration
│
├── docs/                     # Documentation
├── justfile                  # Task definitions
└── CLAUDE.md                 # Claude Code instructions
```

## Available Tasks

```bash
just build              # Build binary to ./bmaduum
just test               # Run all tests
just test-verbose       # Run tests with verbose output
just test-pkg ./internal/claude  # Test specific package
just test-coverage      # Generate coverage.html
just lint               # Run golangci-lint
just fmt                # Format code
just vet                # Run go vet
just check              # Run fmt, vet, and test
just run --help         # Build and run with arguments
```

## Testing

### Key Mocks

- **`claude.MockExecutor`** - Predefined events, exit codes, captures prompts
- **`cli.MockWorkflowRunner`** - Records executed workflows, configurable failures
- **`cli.MockStatusWriter`** - Records status updates
- **`cli.MockBmadHelpFallback`** - Configurable workflow/error responses
- **`bmadhelp.MockFallback`** - Tracks calls with story key and status
- **`output.NewPrinterWithWriter(buf)`** - Captures terminal output

### Test Patterns

```go
// Table-driven tests
func TestGetWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        status  status.Status
        want    string
        wantErr error
    }{
        {"backlog routes to create-story", status.StatusBacklog, "create-story", nil},
        {"done returns error", status.StatusDone, "", router.ErrStoryComplete},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* ... */ })
    }
}
```

## Adding a New Workflow

1. Add to `config/workflows.yaml` with both `slash_command` and `prompt_template`
2. Add to `DefaultConfig()` in `internal/config/types.go`
3. Add routing in `internal/router/router.go` (or rely on manifest discovery)
4. Add tests

## Adding a Configuration Option

1. Add field to appropriate struct in `internal/config/types.go`
2. Add default value in `DefaultConfig()`
3. Add YAML key in `config/workflows.yaml`
4. Add test in `config_test.go`

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (`just fmt`)
- Run `go vet` for static analysis (`just vet`)
- Table-driven tests with descriptive names
- Interfaces end in `-er` (Executor, Printer) or describe capability
