# BMAD Automate - Full Story Lifecycle Automation

## What This Is

A CLI tool that orchestrates Claude CLI to run complete automated development workflows. Stories run through their full lifecycle (create→dev→review→commit) automatically, with error recovery, dry-run mode, and step progress visibility. Fully documented with go doc comments and runnable examples.

## Core Value

Fully automated story lifecycle execution — run a story once and watch it complete from current status to done, with auto-status updates and error recovery.

## Current State (v1.2)

**Shipped:** 2026-01-09

Full story lifecycle execution and comprehensive documentation:

- `run <story>` — Executes complete lifecycle from current status to done
- `queue <story>...` — Full lifecycle for each story, skips done, fails fast
- `epic <epic-id>` — Full lifecycle for all stories in an epic
- `--dry-run` — Preview workflow sequence without execution
- Error recovery — State saved on failure, resumable
- Go doc comments — All internal packages documented with examples
- CLI Cookbook — 14 recipe-style examples in docs/examples/

Tech stack: Go, Cobra, Viper, yaml.v3
Codebase: 8,448 LOC Go

## Requirements

### Validated

- ✓ CLI command structure with Cobra — existing
- ✓ Configuration via Viper with YAML and env vars — existing
- ✓ Claude CLI subprocess execution with streaming JSON — existing
- ✓ Event-driven output parsing — existing
- ✓ Terminal formatting with Lipgloss — existing
- ✓ Commands: create-story, dev-story, code-review, git-commit, run, queue, epic, raw — v1.0
- ✓ Interface-based design for testability (Executor, Printer) — existing
- ✓ Go template expansion for prompts — existing
- ✓ Status-based workflow routing from sprint-status.yaml — v1.0
- ✓ Run command auto-routing based on status — v1.0
- ✓ Queue command with status routing and done-skip — v1.0
- ✓ Epic command for batch execution with numeric sorting — v1.0
- ✓ Fail-fast on story failure — v1.0
- ✓ Full story lifecycle execution (create→dev→review→commit) — v1.1
- ✓ Auto-status updates after each workflow step — v1.1
- ✓ Lifecycle orchestration with interface-based DI — v1.1
- ✓ State persistence for error recovery — v1.1
- ✓ Dry-run mode for workflow preview — v1.1
- ✓ Step progress visibility with callbacks — v1.1
- ✓ Go doc comments for all internal packages — v1.2
- ✓ Runnable Example functions in doc_test.go — v1.2
- ✓ Package-level doc.go files with overviews — v1.2
- ✓ Updated docs folder for v1.1 features — v1.2
- ✓ CLI Cookbook with recipe-style examples — v1.2

### Active

(None currently — v1.2 milestone complete)

### Out of Scope

- Manual workflow override flag — status always determines workflow
- Parallel story execution — sequential only
- Epic status auto-transitions — only story-level routing

## Context

**Sprint Status File:**

- Always located at `_bmad-output/implementation-artifacts/sprint-status.yaml`
- YAML format with `development_status` section
- Story keys follow pattern: `{epic#}-{story#}-{description}` (e.g., `7-1-define-schema`)
- Statuses: `backlog`, `ready-for-dev`, `in-progress`, `review`, `done`

**Workflow Mapping:**
| Status | Workflow |
|--------|----------|
| `backlog` | `/bmad:bmm:workflows:create-story` |
| `ready-for-dev` | `/bmad:bmm:workflows:dev-story` |
| `in-progress` | `/bmad:bmm:workflows:dev-story` |
| `review` | `/bmad:bmm:workflows:code-review` |

## Constraints

- **Tech Stack**: Go with existing Cobra/Viper patterns
- **File Location**: Sprint status always at `_bmad-output/implementation-artifacts/sprint-status.yaml`
- **Claude CLI**: Requires Claude CLI installed and in PATH

## Key Decisions

| Decision                              | Rationale                                             | Outcome |
| ------------------------------------- | ----------------------------------------------------- | ------- |
| Auto-detect only, no manual override  | Simplicity — status is source of truth                | ✓ Good  |
| Stop on first failure in epic         | Allows investigation before continuing                | ✓ Good  |
| Sequential execution only             | Stories may have dependencies                         | ✓ Good  |
| yaml.v3 instead of Viper for status   | Simpler for single file with known structure          | ✓ Good  |
| Package-level router function         | Pure mapping with no state needed                     | ✓ Good  |
| StatusReader injected via App struct  | Testability — allows mock injection                   | ✓ Good  |
| Done stories skipped in queue         | Allows mixed-status batches without failure           | ✓ Good  |
| Epic reuses QueueRunner               | DRY — inherits all routing, skip, and fail-fast logic | ✓ Good  |
| Interface-based DI for lifecycle      | WorkflowRunner, StatusReader, StatusWriter interfaces | ✓ Good  |
| Fail-fast on workflow or status fail  | Allows investigation before continuing                | ✓ Good  |
| Manager pattern for state             | Testable file operations with injected directory      | ✓ Good  |
| Atomic writes for state files         | Temp file + rename prevents corruption                | ✓ Good  |
| SetProgressCallback (not constructor) | Optional callback keeps NewExecutor signature simple  | ✓ Good  |

| Used doc_test.go instead of doc.go | Go requires \_test.go for external test packages | ✓ Good |
| Recipe format for CLI cookbook | Brief, scannable, copy-paste-ready | ✓ Good |

---

_Last updated: 2026-01-09 after v1.2 milestone_
