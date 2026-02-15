This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
just build              # Build binary to ./bmaduum
just test               # Run all tests
just test-verbose       # Run tests with verbose output
just test-pkg ./internal/claude  # Test specific package
just test-coverage      # Generate coverage.html
just lint               # Run golangci-lint
just check              # Run fmt, vet, and test
just run --help         # Build and run with arguments
just release-snapshot   # Build release locally (snapshot)
just release            # Full release with GoReleaser
```

## Architecture

This is a CLI tool that orchestrates Claude CLI to run automated development workflows. It spawns Claude as a subprocess, parses its streaming JSON output, and displays formatted results.

### Package Dependencies

```
cmd/bmaduum/main.go
         |
         v
    internal/cli (Cobra commands)
         |
         ├──> internal/workflow (orchestration)
         |         |
         |         ├──> internal/claude (Claude execution + JSON parsing)
         |         |
         |         └──> internal/output (terminal formatting)
         |
         ├──> internal/lifecycle (lifecycle orchestration)
         |         |
         |         ├──> internal/router (GetLifecycle for step sequences)
         |         |
         |         └──> internal/state (execution state persistence)
         |
         ├──> internal/ratelimit (rate limit detection from stderr)
         |
         ├──> internal/status (sprint status reading)
         |
         ├──> internal/router (workflow routing)
         |
         └──> internal/config (Viper configuration)
```

### Key Interfaces for Testing

- **`claude.Executor`** - Interface for running Claude CLI. Use `MockExecutor` in tests to avoid spawning real processes.
- **`output.Printer`** - Interface for terminal output. Use `NewPrinterWithWriter(buf)` to capture output in tests.

### Data Flow

1. CLI command receives story key
2. `config.Config.GetPrompt()` expands Go template with `{{.StoryKey}}`
3. `workflow.Runner` calls `claude.Executor.ExecuteWithResult()`
4. `claude.Parser` reads streaming JSON, emits `Event` structs
5. `output.Printer` formats and displays events
6. Lifecycle executor runs multiple workflows with state persistence

### Configuration

Workflow prompts are in `config/workflows.yaml` using Go templates. Config loads via Viper with env var overrides (`BMADUUM_` prefix).

### Claude CLI Integration

The executor always passes `--dangerously-skip-permissions` and `--output-format stream-json`. Each JSON line from stdout is parsed into `StreamEvent` structs, then converted to the higher-level `Event` type with convenience methods (`IsText()`, `IsToolUse()`, `IsToolResult()`).

### Rate Limit Detection

The `internal/ratelimit` package provides rate limit detection from Claude CLI's stderr output:

- `Detector` identifies rate limit error messages and extracts reset times
- `State` provides thread-safe rate limit state management
- Used in conjunction with `--auto-retry` flag for automatic retry with intelligent wait times

### State Persistence

The `internal/state` package enables resume functionality when lifecycle execution fails:

- State is saved to `.bmad-state.json` in the working directory
- Atomic writes (temp file + rename) ensure crash safety
- Automatically cleared on successful completion
- Tracks story key, step index, total steps, and start status

---

## BMAD-METHOD v6 Integration

bmaduum was originally built against an earlier BMAD-METHOD version. The goal is to update it to work properly with BMAD-METHOD v6.0.0 Beta 8+ rather than bypassing its workflow engine with raw prompts.

### The Core Problem

bmaduum currently constructs its own prompts and sends them directly to Claude CLI, effectively reimplementing workflow logic that BMAD v6 agents already handle. V6 introduced sharded step-based execution, specialized agents, slash commands, and a module system — none of which bmaduum leverages.

### BMAD v6 File Structure

```
project-root/
├── _bmad/                          # BMAD installation root
│   ├── _cfg/                       # Configuration
│   │   ├── workflow-manifest.csv   # All available workflows, phases, commands
│   │   └── manifest.yaml           # Installed modules, agents, tools
│   ├── _config/
│   │   ├── manifest.yaml           # Installation manifest
│   │   └── custom/                 # Custom module cache
│   ├── _memory/                    # Agent sidecar content
│   │   └── {agent-name}-sidecar/
│   ├── bmm/                        # BMad Method module
│   │   ├── agents/                 # Agent definitions (Dev, PM, SM, QA, etc.)
│   │   ├── workflows/              # Phase 1-4 workflows (sharded step-*.md files)
│   │   └── docs/
│   └── modules/                    # External modules (TEA, BMGD, CIS, etc.)
├── _bmad-output/
│   ├── planning-artifacts/         # PRDs, architecture docs, epics
│   └── implementation-artifacts/
│       └── sprint-status.yaml      # Source of truth for story status
├── .claude/
│   └── commands/
│       └── bmad-*.md               # Slash commands for Claude Code
```

### Key Integration Points

**Slash commands** are the primary invocation method in v6. Instead of constructing prompts, bmaduum should invoke these:
- `/create-story` — SM agent creates a story from epics
- `/dev-story` — Dev agent implements a story
- `/code-review` — QA/SDET agent reviews code
- `/sprint-status` — Check sprint progress
- `/bmad-help` — Context-aware routing (what to do next)

**Workflow manifest** (`_bmad/_cfg/workflow-manifest.csv`) catalogs all installed workflows with their phase, agent, and slash command. Use this for dynamic discovery instead of hardcoding the workflow chain.

**Sprint status** file may be at `_bmad-output/implementation-artifacts/sprint-status.yaml`. V6 also supports a `project_key` system for file-system based tracking. The path should be discoverable from config, not hardcoded.

**Agent activation**: Each workflow step should run under the correct BMAD agent persona. In Claude Code, slash commands already route to the right agent, so using slash commands solves this automatically.

**Sharded workflows**: V6 workflows use step-*.md files loaded sequentially by a workflow.xml meta-executor. The AI only sees the current step. bmaduum should NOT try to replicate this — invoking the slash command lets BMAD handle it.

### Migration Plan (in priority order)

#### Phase 1: Slash Command Invocation
The highest-impact change. Replace raw prompt construction with slash command invocation:

```go
// OLD: bmaduum constructs a long prompt
prompt := fmt.Sprintf("You are a developer. Implement story %s...", storyID)
exec.Run("claude", "--dangerously-skip-permissions", "-p", prompt)

// NEW: let BMAD's workflow engine handle it
exec.Run("claude", "--dangerously-skip-permissions", "-p",
    fmt.Sprintf("/dev-story %s", storyID))
```

This single change activates the full v6 engine — agent personas, step-by-step execution, progressive disclosure, checklists.

Update `config/workflows.yaml` templates accordingly. The prompts become thin wrappers around slash commands rather than full workflow instructions.

**Files to change**: `config/workflows.yaml`, `internal/workflow/`, `internal/config/`

#### Phase 2: Sprint Status Path Update
Update `internal/status/` to:
1. Check `_bmad-output/implementation-artifacts/sprint-status.yaml` (v6 path)
2. Fall back to the old path for backward compatibility
3. Allow override via config (`BMADUUM_SPRINT_STATUS_PATH`)

**Files to change**: `internal/status/`, `internal/config/`

#### Phase 3: Workflow Manifest Discovery
Add a new `internal/manifest/` package that reads `_bmad/_cfg/workflow-manifest.csv` to dynamically discover available workflows. Replace the hardcoded status→workflow chain in `internal/router/` with manifest-driven routing.

**Files to change**: new `internal/manifest/`, `internal/router/`

#### Phase 4: Module Awareness
Read `_bmad/_config/manifest.yaml` to discover installed modules. If TEA/SDET is installed, the story lifecycle could include a test-automation step. Surface module info in `--dry-run` output.

**Files to change**: `internal/manifest/`, `internal/lifecycle/`

#### Phase 5: bmad-help Integration
Use `/bmad-help` as a routing oracle for edge cases instead of maintaining hardcoded fallback logic. When bmaduum can't determine the next step from sprint-status alone, delegate to bmad-help.

### Testing Strategy

- All existing tests must pass after changes (`just test`)
- Mock the Claude executor interface — don't spawn real Claude processes in tests
- Add integration test fixtures with sample v6 `_bmad/` directory structures
- Test backward compatibility with projects that don't have v6 layout
- Use `--dry-run` extensively to verify correct slash commands are generated before live testing

### Important Constraints

- Keep backward compatibility with pre-v6 BMAD projects where possible
- Don't modify BMAD-METHOD files — bmaduum is a consumer, not a modifier
- Preserve the `--dangerously-skip-permissions` and `--output-format stream-json` flags
- Maintain the rate limit detection and auto-retry functionality
- State persistence (`.bmad-state.json`) should continue to work for crash recovery
