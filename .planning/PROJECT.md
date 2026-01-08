# BMAD Automate - Status-Based Workflow Routing

## What This Is

A CLI tool that orchestrates Claude CLI to run automated development workflows. Currently being enhanced to automatically detect story status from sprint-status.yaml and route to the appropriate workflow, plus a new `epic` command to batch-run all stories in an epic.

## Core Value

Eliminate manual workflow selection by automatically routing stories to the correct workflow based on their status in sprint-status.yaml.

## Requirements

### Validated

- ✓ CLI command structure with Cobra — existing
- ✓ Configuration via Viper with YAML and env vars — existing
- ✓ Claude CLI subprocess execution with streaming JSON — existing
- ✓ Event-driven output parsing — existing
- ✓ Terminal formatting with Lipgloss — existing
- ✓ Commands: create-story, dev-story, code-review, git-commit, run, queue, raw — existing
- ✓ Interface-based design for testability (Executor, Printer) — existing
- ✓ Go template expansion for prompts — existing

### Active

- [ ] Status-based workflow routing: CLI reads story status from `_bmad-output/implementation-artifacts/sprint-status.yaml` and routes to correct workflow
  - `backlog` → `create-story`
  - `ready-for-dev` / `in-progress` → `dev-story`
  - `review` → `code-review`
- [ ] Apply routing to all story execution: `run`, `queue`, and new `epic` commands
- [ ] New `epic` command: `bmad-automate epic <epic-id>` runs all non-done stories for an epic in numeric order
- [ ] Stop immediately on story failure during epic execution

### Out of Scope

- Manual workflow override flag — status always determines workflow
- Updating sprint-status.yaml after story completion — read-only access
- Parallel story execution — sequential only
- Epic status auto-transitions — only story-level routing

## Context

**Sprint Status File:**

- Always located at `_bmad-output/implementation-artifacts/sprint-status.yaml`
- YAML format with `development_status` section
- Story keys follow pattern: `{epic#}-{story#}-{description}` (e.g., `7-1-define-schema`)
- Statuses: `backlog`, `ready-for-dev`, `in-progress`, `review`, `done`

**Existing Architecture:**

- Layered CLI with dependency injection
- Stateless execution model
- Commands use `RunE` pattern for testable exit codes
- Workflows defined in `config/workflows.yaml` using Go templates

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

| Decision                             | Rationale                                         | Outcome   |
| ------------------------------------ | ------------------------------------------------- | --------- |
| Auto-detect only, no manual override | Simplicity — status is source of truth            | — Pending |
| Stop on first failure in epic        | Allows investigation before continuing            | — Pending |
| Sequential execution only            | Stories may have dependencies                     | — Pending |
| Read-only sprint-status.yaml access  | Separation of concerns — status managed elsewhere | — Pending |

---

_Last updated: 2026-01-08 after initialization_
