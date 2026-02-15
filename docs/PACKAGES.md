# Package Documentation

API reference for all internal packages in `bmaduum`.

## Package Overview

| Package | Location | Purpose |
|---------|----------|---------|
| [cli](#cli) | `internal/cli/` | CLI commands, dependency injection, error handling |
| [claude](#claude) | `internal/claude/` | Claude CLI execution and JSON parsing |
| [config](#config) | `internal/config/` | Configuration loading and template expansion |
| [output](#output) | `internal/output/` | Terminal formatting and styling |
| [workflow](#workflow) | `internal/workflow/` | Single workflow execution |
| [lifecycle](#lifecycle) | `internal/lifecycle/` | Story lifecycle orchestration |
| [router](#router) | `internal/router/` | Status-to-workflow routing (hardcoded or manifest-driven) |
| [manifest](#manifest) | `internal/manifest/` | Workflow manifest CSV and module YAML parsing |
| [bmadhelp](#bmadhelp) | `internal/bmadhelp/` | bmad-help fallback for unknown statuses |
| [status](#status) | `internal/status/` | Sprint status file reading with path discovery |
| [state](#state) | `internal/state/` | Lifecycle state persistence for resume |
| [ratelimit](#ratelimit) | `internal/ratelimit/` | Rate limit detection from Claude stderr |

---

## cli

**Package:** `internal/cli`

Command-line interface implementation using Cobra framework.

### App

Dependency injection container holding all application dependencies.

```go
type App struct {
    Config       *config.Config
    Executor     claude.Executor
    Printer      core.Printer
    Runner       WorkflowRunner
    StatusReader StatusReader
    StatusWriter StatusWriter
    Router       *router.Router             // Manifest-driven or hardcoded defaults
    Modules      *manifest.ModuleManifest   // nil if no module manifest found
    BmadHelp     lifecycle.BmadHelpFallback // nil disables fallback
}
```

### Key Functions

```go
func NewApp(cfg *config.Config) *App          // Wire up all production dependencies
func NewRootCommand(app *App) *cobra.Command   // Create command tree
func RunWithConfig(cfg *config.Config) ExecuteResult  // Testable entry point
func Execute()                                 // main() entry point (calls os.Exit)
```

`NewApp` loads the workflow manifest, module manifest, and wires up the bmad-help fallback automatically.

---

## claude

**Package:** `internal/claude`

Claude CLI subprocess execution and JSON stream parsing.

### Executor Interface

```go
type Executor interface {
    Execute(ctx context.Context, prompt string) (<-chan Event, error)
    ExecuteWithResult(ctx context.Context, prompt string, handler EventHandler, model string) (int, error)
}
```

`MockExecutor` provides a test implementation with `Events`, `ExitCode`, `Error`, and `RecordedPrompts` fields.

### Event

Parsed event from Claude's streaming JSON output with convenience methods:

```go
func (e Event) IsText() bool
func (e Event) IsToolUse() bool
func (e Event) IsToolResult() bool
```

---

## config

**Package:** `internal/config`

Configuration loading via Viper with Go template expansion.

### Config

```go
type Config struct {
    UseSlashCommands bool                       // true = v6 slash commands, false = legacy templates
    Workflows        map[string]WorkflowConfig  // Workflow definitions
    StatusPath       string                     // Explicit sprint-status.yaml path (auto-discovered if empty)
    Claude           ClaudeConfig               // Claude CLI settings
    Output           OutputConfig               // Terminal output settings
}
```

### WorkflowConfig

```go
type WorkflowConfig struct {
    SlashCommand   string  // v6 template: "/dev-story {{.StoryKey}}"
    PromptTemplate string  // Legacy template: "/bmad-bmm-dev-story - ..."
    Model          string  // Optional model override (e.g., "opus", "sonnet")
}
```

### GetPrompt

```go
func (c *Config) GetPrompt(workflowName, storyKey string) (string, error)
```

Returns the expanded prompt. Uses `SlashCommand` when `UseSlashCommands` is true, `PromptTemplate` when false. Falls back to the other template if the selected one is empty.

### GetModel

```go
func (c *Config) GetModel(workflowName string) string
```

Returns the model override for a workflow, or empty string for default.

---

## workflow

**Package:** `internal/workflow`

Single workflow execution using Claude CLI.

### Runner

```go
type Runner struct { /* ... */ }

func NewRunner(executor claude.Executor, printer core.Printer, cfg *config.Config) *Runner
func (r *Runner) RunSingle(ctx context.Context, workflowName, storyKey string) int
func (r *Runner) RunRaw(ctx context.Context, prompt string) int
func (r *Runner) SetOperation(operation string)  // Set progress bar context
```

`RunSingle` calls `config.GetPrompt()` to expand the slash command template, then executes Claude CLI with streaming output.

---

## lifecycle

**Package:** `internal/lifecycle`

Story lifecycle orchestration from current status to done.

### Executor

```go
type Executor struct { /* ... */ }

func NewExecutor(runner WorkflowRunner, reader StatusReader, writer StatusWriter) *Executor
func (e *Executor) SetRouter(r *router.Router)
func (e *Executor) SetBmadHelp(fb BmadHelpFallback)
func (e *Executor) Execute(ctx context.Context, storyKey string) error
func (e *Executor) GetSteps(storyKey string) ([]router.LifecycleStep, error)
```

`Execute` looks up the story status, determines remaining steps via the router, and runs each workflow in sequence. After each success, it updates the story status.

When the router returns `ErrUnknownStatus` and bmad-help is configured, the executor invokes `/bmad-help` to get a single workflow recommendation, executes it, then re-reads the status and continues. This is depth-limited to 3 recursive calls.

### BmadHelpFallback

```go
type BmadHelpFallback interface {
    ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (workflow string, nextStatus status.Status, err error)
}
```

---

## router

**Package:** `internal/router`

Status-to-workflow routing with support for hardcoded defaults and manifest-driven routing.

### Router

```go
type Router struct { /* ... */ }

func NewRouter() *Router                                   // Hardcoded defaults
func NewRouterFromManifest(m *manifest.Manifest) *Router   // Manifest-driven
func (r *Router) GetWorkflow(s status.Status) (string, error)
func (r *Router) GetLifecycle(s status.Status) ([]LifecycleStep, error)
func (r *Router) InsertStepAfter(after, workflow string, nextStatus status.Status)
```

### LifecycleStep

```go
type LifecycleStep struct {
    Workflow   string
    NextStatus status.Status
    Model      string
}
```

### Sentinel Errors

```go
var ErrStoryComplete = errors.New("story is complete, no workflow needed")
var ErrUnknownStatus = errors.New("unknown status value")
```

Package-level `GetWorkflow()` and `GetLifecycle()` functions are available as backward-compatible wrappers using a default hardcoded router.

---

## manifest

**Package:** `internal/manifest`

BMAD v6 file discovery for workflow manifests and module manifests.

### Workflow Manifest

Reads `_bmad/_cfg/workflow-manifest.csv` to discover available workflows.

```go
type Manifest struct { /* ... */ }

func ReadFromFile(path string) (*Manifest, error)
func (m *Manifest) HasWorkflow(name string) bool
func (m *Manifest) GetEntriesForStatus(status string) []WorkflowEntry
```

### Module Manifest

Reads `_bmad/_config/manifest.yaml` to discover installed modules.

```go
type ModuleManifest struct { /* ... */ }

func ReadModulesFromFile(path string) (*ModuleManifest, error)
func (m *ModuleManifest) HasModule(name string) bool
func (m *ModuleManifest) Names() []string
```

---

## bmadhelp

**Package:** `internal/bmadhelp`

Last-resort fallback for unknown statuses via `/bmad-help` invocation.

```go
type ClaudeFallback struct { /* ... */ }

func NewClaudeFallback(executor claude.Executor) *ClaudeFallback
func (c *ClaudeFallback) ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (string, status.Status, error)
```

Invokes `/bmad-help` via Claude CLI, parses the response for known workflow names (create-story, dev-story, code-review, test-automation, git-commit), and returns the recommended workflow and expected next status.

`ParseResponse(response string) (*Recommendation, error)` extracts workflow names from free-form text (case-insensitive, first match wins).

`MockFallback` is available for testing.

---

## status

**Package:** `internal/status`

Sprint status file reading with v6/legacy path discovery.

### Path Discovery

```go
func ResolvePath(configuredPath string) string
```

Returns the sprint-status.yaml path using priority: `BMADUUM_SPRINT_STATUS_PATH` env var > `configuredPath` > v6 path > legacy path.

### Reader / Writer

```go
func NewReader(basePath string) *Reader               // Auto-discovers path
func NewReaderWithPath(basePath, configuredPath string) *Reader  // Uses configured path
func NewWriter(basePath string) *Writer
func NewWriterWithPath(basePath, configuredPath string) *Writer

func (r *Reader) GetStoryStatus(storyKey string) (Status, error)
func (r *Reader) GetEpicStories(epicID string) ([]string, error)
func (r *Reader) GetAllEpics() ([]string, error)
func (w *Writer) UpdateStatus(storyKey string, newStatus Status) error  // Atomic write
```

---

## state

**Package:** `internal/state`

Lifecycle state persistence for resume functionality. State is saved to `.bmad-state.json` using atomic writes (temp file + rename).

```go
func NewManager(dir string) *Manager
func (m *Manager) Save(state State) error    // Atomic write
func (m *Manager) Load() (State, error)      // Returns ErrNoState if absent
func (m *Manager) Clear() error              // Idempotent
```

---

## ratelimit

**Package:** `internal/ratelimit`

Rate limit detection from Claude CLI stderr output.

```go
func NewDetector() *Detector
func (d *Detector) CheckLine(line string) Info    // Detect rate limit messages
func (d *Detector) WaitTime(info Info) time.Duration  // Calculate wait time
```

Used with `--auto-retry` flag for automatic retry with intelligent wait times.
