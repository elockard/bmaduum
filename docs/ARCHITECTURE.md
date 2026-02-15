# Architecture Documentation

## System Overview

`bmaduum` is a Go CLI tool that orchestrates Claude AI to automate BMAD-METHOD development workflows. It invokes BMAD v6 slash commands via Claude CLI as a subprocess, parses streaming JSON output, and displays formatted results.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              bmaduum                                        │
│                                                                             │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌────────────┐  │
│  │  CLI Layer  │───>│  Lifecycle   │───>│   Workflow  │───>│   Claude   │  │
│  │   (Cobra)   │    │  (Executor)  │    │   (Runner)  │    │ (Executor) │  │
│  └─────────────┘    └──────────────┘    └─────────────┘    └────────────┘  │
│         │                  │                   │                  │         │
│         v                  v                   v                  v         │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌────────────┐  │
│  │   Config    │    │   Router     │    │   Status    │    │   Output   │  │
│  │   (Viper)   │    │  (Manifest)  │    │  (Reader)   │    │  (Printer) │  │
│  └─────────────┘    └──────────────┘    └─────────────┘    └────────────┘  │
│         │                  │                                               │
│         v                  v                                               │
│  ┌─────────────┐    ┌──────────────┐                                       │
│  │  Manifest   │    │  BmadHelp   │                                       │
│  │ (CSV/YAML)  │    │ (Fallback)  │                                       │
│  └─────────────┘    └──────────────┘                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      v
                          ┌───────────────────────┐
                          │      Claude CLI       │
                          │  (External Process)   │
                          └───────────────────────┘
```

## Architecture Pattern

**Pattern:** Layered CLI Application with Dependency Injection

**Key Characteristics:**

- Single executable with subcommands
- Subprocess orchestration (wraps Claude CLI)
- Slash command invocation (delegates workflow logic to BMAD v6 engine)
- Lifecycle-driven execution with state persistence for resume
- Event-driven streaming output
- Interface-based design for testability

## Package Dependencies

```
cmd/bmaduum/main.go
         │
         v
    internal/cli (Cobra commands)
         │
         ├──> internal/lifecycle (lifecycle orchestration)
         │         │
         │         ├──> internal/router (GetLifecycle for step sequences)
         │         │
         │         └──> internal/workflow (WorkflowRunner for execution)
         │
         ├──> internal/bmadhelp (bmad-help fallback for unknown statuses)
         │         │
         │         └──> internal/claude (Claude execution)
         │
         ├──> internal/manifest (workflow CSV + module YAML discovery)
         │
         ├──> internal/state (execution state persistence)
         │
         ├──> internal/workflow (single workflow orchestration)
         │         │
         │         ├──> internal/claude (Claude execution + JSON parsing)
         │         │
         │         ├──> internal/output (terminal formatting)
         │         │
         │         └──> internal/config (configuration + prompt expansion)
         │
         ├──> internal/status (sprint status reading with path discovery)
         │
         ├──> internal/router (workflow routing: hardcoded or manifest-driven)
         │
         └──> internal/config (Viper configuration)
```

## Data Flow

### Slash Command Invocation

When `use_slash_commands` is `true` (default), the flow is:

```
1. CLI receives: bmaduum story 6-1-setup
2. config.GetPrompt("dev-story", "6-1-setup") → "/dev-story 6-1-setup"
3. Claude CLI executes: claude --dangerously-skip-permissions -p "/dev-story 6-1-setup" ...
4. BMAD v6 engine activates the correct agent and workflow steps
5. Streaming JSON events → parsed → formatted terminal output
```

When `use_slash_commands` is `false`, legacy prompt templates are used instead. The template is sent as the prompt, and the BMAD v6 engine is bypassed.

### Lifecycle Execution Flow

```
┌────────────────────────────────────────────────────────────────────────────┐
│  bmaduum story 6-1-setup                                                   │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    v
┌────────────────────────────────────────────────────────────────────────────┐
│  1. Status Reader                                                          │
│     - Auto-discovers sprint-status.yaml (v6 path > legacy path)            │
│     - Get status for 6-1-setup: "backlog"                                  │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    v
┌────────────────────────────────────────────────────────────────────────────┐
│  2. Router                                                                 │
│     - If manifest exists: manifest-driven routing                          │
│     - Otherwise: hardcoded routing table                                   │
│     - router.GetLifecycle("backlog") → 4 steps                             │
│                                                                            │
│     Unknown status? → bmad-help fallback (if enabled)                      │
│       Invokes /bmad-help via Claude CLI                                    │
│       Parses response to determine next workflow                           │
│       Depth-limited to 3 recursive calls                                   │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    v
┌────────────────────────────────────────────────────────────────────────────┐
│  3. Execute Steps Loop                                                     │
│                                                                            │
│     for each step:                                                         │
│       a. Call progressCallback(stepIndex, totalSteps, workflow)             │
│       b. runner.RunSingle(ctx, workflow, storyKey)                         │
│          → config.GetPrompt() expands slash command template               │
│          → executor.ExecuteWithResult() spawns Claude CLI                  │
│       c. If exit code != 0 → return error (fail-fast)                      │
│       d. statusWriter.UpdateStatus(storyKey, nextStatus)                   │
│                                                                            │
│     Success: all steps completed, story is done                            │
│     Failure: stops at first error                                          │
└────────────────────────────────────────────────────────────────────────────┘
```

### Module-Aware Lifecycle

When BMAD modules are installed:

```
1. On startup, NewApp() reads _bmad/_config/manifest.yaml
2. If SDET or TEA module found → router.InsertStepAfter("code-review", "test-automation", done)
3. Lifecycle becomes: ... → code-review → test-automation → git-commit → done
4. Module info shown in --dry-run output
```

## Key Interfaces

### Executor Interface

```go
type Executor interface {
    Execute(ctx context.Context, prompt string) (<-chan Event, error)
    ExecuteWithResult(ctx context.Context, prompt string, handler EventHandler, model string) (int, error)
}
```

### Lifecycle Interfaces

```go
type WorkflowRunner interface {
    RunSingle(ctx context.Context, workflowName, storyKey string) int
}

type StatusReader interface {
    GetStoryStatus(storyKey string) (status.Status, error)
}

type StatusWriter interface {
    UpdateStatus(storyKey string, newStatus status.Status) error
}

type BmadHelpFallback interface {
    ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (workflow string, nextStatus status.Status, err error)
}
```

## App Dependency Container

```go
type App struct {
    Config       *config.Config
    Executor     claude.Executor
    Printer      core.Printer
    Runner       WorkflowRunner
    StatusReader StatusReader
    StatusWriter StatusWriter
    Router       *router.Router        // Manifest-driven or hardcoded
    Modules      *manifest.ModuleManifest  // nil if no module manifest
    BmadHelp     lifecycle.BmadHelpFallback // nil disables fallback
}
```

## Configuration Loading

```
Priority (highest to lowest):
  1. Environment variables (BMADUUM_*)
  2. BMADUUM_CONFIG_PATH explicit file
  3. ~/.config/bmaduum/workflows.yaml (platform-standard)
  4. ./config/workflows.yaml (legacy)
  5. ./workflows.yaml (legacy)
  6. Built-in DefaultConfig()

Sprint Status Path:
  1. BMADUUM_SPRINT_STATUS_PATH env var
  2. status_path config value
  3. _bmad-output/implementation-artifacts/sprint-status.yaml (v6)
  4. sprint-status.yaml (legacy)
```

## Design Principles

1. **Slash Command Delegation** - Let BMAD v6 handle agent personas and workflow logic
2. **Dependency Injection** - All dependencies injected via App struct
3. **Interface Segregation** - Small, focused interfaces (Executor, Printer, WorkflowRunner)
4. **Backward Compatibility** - Dual-mode prompts (slash commands + legacy templates)
5. **Dynamic Discovery** - Manifest-driven routing and module-aware lifecycle
6. **Graceful Fallback** - bmad-help for unknown statuses, auto-discovery for paths
7. **Fail-Fast Execution** - Stop immediately on error, save state for resume
8. **Testability First** - Interfaces and mocks for isolated testing
