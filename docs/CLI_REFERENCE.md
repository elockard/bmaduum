# CLI Reference

Complete command-line interface reference for `bmaduum`.

## Synopsis

```
bmaduum [command] [arguments] [flags]
```

## Description

BMAD Automation CLI orchestrates Claude AI to run development workflows. By default, it invokes BMAD v6 slash commands (`/create-story`, `/dev-story`, etc.), letting the BMAD workflow engine handle agent personas and step-by-step execution.

## Global Behavior

All commands:

- Load configuration from `config/workflows.yaml` (or `BMADUUM_CONFIG_PATH`)
- Execute Claude CLI with `--dangerously-skip-permissions` and `--output-format stream-json`
- Display styled terminal output with progress indicators
- Return appropriate exit codes (0 for success, non-zero for failure)

---

## Commands

### story

Run full lifecycle for one or more stories from their current status to done.

**Usage:**

```bash
bmaduum story [--dry-run] [--auto-retry] [--no-bmad-help] <story-key> [story-key...]
```

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| story-key | Yes (1+) | One or more story identifiers |

**Flags:**
| Flag | Description |
|------|-------------|
| `--dry-run` | Preview workflow sequence without execution |
| `--auto-retry` | Automatically retry on rate limit errors |
| `--no-bmad-help` | Disable bmad-help fallback for unknown statuses |

**Examples:**

```bash
bmaduum story 6-1-setup-project
bmaduum story 6-1-setup 6-2-auth 6-3-tests
bmaduum story --dry-run 6-1-setup 6-2-auth
```

**Behavior:**

1. Processes each story through its **full lifecycle** to completion
2. Auto-updates status after each successful workflow step
3. Skips stories with status `done`
4. Stops on first failure
5. For unrecognized statuses, invokes `/bmad-help` fallback (unless `--no-bmad-help`)

**Lifecycle Routing:**

| Story Status    | Remaining Lifecycle                                            |
| --------------- | -------------------------------------------------------------- |
| `backlog`       | create-story -> dev-story -> code-review -> git-commit -> done |
| `ready-for-dev` | dev-story -> code-review -> git-commit -> done                 |
| `in-progress`   | dev-story -> code-review -> git-commit -> done                 |
| `review`        | code-review -> git-commit -> done                              |
| `done`          | No action (story already complete)                             |

If a workflow manifest is found at `_bmad/_cfg/workflow-manifest.csv`, routing is driven by the manifest instead of the hardcoded table above.

If SDET or TEA modules are installed (via `_bmad/_config/manifest.yaml`), `test-automation` is automatically inserted after `code-review`.

---

### epic

Run full lifecycle for all stories in one or more epics, or all active epics.

**Usage:**

```bash
bmaduum epic [--dry-run] [--auto-retry] [--no-bmad-help] <epic-id>|all [epic-id...]
```

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| epic-id | Yes (1+) | One or more epic identifiers, or `all` for all active epics |

**Flags:**
| Flag | Description |
|------|-------------|
| `--dry-run` | Preview workflow sequence without execution |
| `--auto-retry` | Automatically retry on rate limit errors |
| `--no-bmad-help` | Disable bmad-help fallback for unknown statuses |

**Examples:**

```bash
bmaduum epic 6
bmaduum epic 2 4 6
bmaduum epic all
bmaduum epic --dry-run all
```

**Story Discovery:**

Stories are discovered from `sprint-status.yaml` using the pattern `{epic-id}-{story-number}-*`. For epic `6`, this matches `6-1-implement-auth`, `6-2-add-dashboard`, etc. Stories are sorted by story number.

---

### workflow (Advanced)

Run individual BMAD workflow steps directly.

**Usage:**

```bash
bmaduum workflow <workflow-name> <story-key>
```

**Available workflows:**
| Subcommand | Description |
|------------|-------------|
| `create-story` | Create a story definition from backlog |
| `dev-story` | Implement a story through development |
| `code-review` | Review code changes for a story |
| `git-commit` | Commit and push changes for a story |

**Flags:**
| Flag | Description |
|------|-------------|
| `--auto-retry` | Automatically retry on rate limit errors |

**Examples:**

```bash
bmaduum workflow create-story 6-1-setup
bmaduum workflow dev-story 6-1-setup
```

**When to use:** Retrying a failed step, running a step out of sequence, or testing workflow prompts. Most users should use `story` or `epic` instead.

---

### raw

Execute an arbitrary prompt with Claude.

**Usage:**

```bash
bmaduum raw <prompt>
```

**Example:**

```bash
bmaduum raw "List all Go files in the project"
```

---

### version

Display version information.

```bash
bmaduum version
```

---

## Exit Codes

| Code | Meaning                                              |
| ---- | ---------------------------------------------------- |
| 0    | Success                                              |
| 1    | General error (config load failure, unknown command) |
| N    | Claude exit code (passed through from Claude CLI)    |

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BMADUUM_CONFIG_PATH` | Path to configuration file | auto-discovered |
| `BMADUUM_CLAUDE_PATH` | Path to claude binary | `claude` |
| `BMADUUM_SPRINT_STATUS_PATH` | Path to sprint-status.yaml | auto-discovered |

---

## Configuration File

Configuration is loaded from (in priority order):

1. `BMADUUM_CONFIG_PATH` environment variable
2. `~/.config/bmaduum/workflows.yaml` (Linux) or platform equivalent
3. `./config/workflows.yaml` (legacy)
4. `./workflows.yaml` (legacy)
5. Built-in defaults

### Example Configuration

```yaml
# Use BMAD v6 slash commands (true) or legacy prompt templates (false)
use_slash_commands: true

# Explicit sprint-status.yaml path (auto-discovered if empty)
# status_path: ""

workflows:
  create-story:
    slash_command: "/create-story {{.StoryKey}}"
    prompt_template: "/bmad-bmm-create-story - Create story: {{.StoryKey}}. Do not ask questions."

  dev-story:
    slash_command: "/dev-story {{.StoryKey}}"
    prompt_template: "/bmad-bmm-dev-story - Work on story: {{.StoryKey}}..."
    # model: opus  # Optional: override Claude model for this workflow

  code-review:
    slash_command: "/code-review {{.StoryKey}}"
    prompt_template: "/bmad-bmm-code-review - Review story: {{.StoryKey}}..."

  git-commit:
    slash_command: "/git-commit {{.StoryKey}}"
    prompt_template: "Commit all changes for story {{.StoryKey}}..."

claude:
  output_format: stream-json
  binary_path: claude

output:
  truncate_lines: 20
  truncate_length: 60
```

### Configuration Options

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `use_slash_commands` | bool | `true` | Use v6 slash commands vs legacy prompt templates |
| `status_path` | string | `""` | Explicit sprint-status.yaml path (auto-discovered if empty) |
| `workflows.<name>.slash_command` | string | | BMAD v6 slash command template |
| `workflows.<name>.prompt_template` | string | | Legacy prompt template |
| `workflows.<name>.model` | string | `""` | Claude model override for this workflow |
| `claude.binary_path` | string | `claude` | Path to Claude CLI binary |
| `claude.output_format` | string | `stream-json` | Claude output format |
| `output.truncate_lines` | int | `20` | Max lines for tool output display |
| `output.truncate_length` | int | `60` | Max chars for command headers |

### Prompt Mode

When `use_slash_commands` is `true` (default), `GetPrompt()` returns the `slash_command` template. When `false`, it returns the `prompt_template`. If the selected template is empty, the other is used as fallback.

### Template Variables

| Variable        | Description                         |
| --------------- | ----------------------------------- |
| `{{.StoryKey}}` | The story key passed to the command |

---

## Sprint Status File

Auto-discovered in priority order:

1. `BMADUUM_SPRINT_STATUS_PATH` environment variable
2. `status_path` config value
3. `_bmad-output/implementation-artifacts/sprint-status.yaml` (v6 path)
4. `sprint-status.yaml` (legacy path)

**Format:**

```yaml
development_status:
  6-1-setup-project: ready-for-dev
  6-2-add-authentication: in-progress
  6-3-fix-bug: review
  6-4-documentation: done
```

**Valid Status Values:**

- `backlog` - Story not yet started
- `ready-for-dev` - Story ready for implementation
- `in-progress` - Story being implemented
- `review` - Story in code review
- `done` - Story complete

---

## BMAD v6 Integration

### Workflow Manifest

If `_bmad/_cfg/workflow-manifest.csv` exists, bmaduum uses it for dynamic workflow routing instead of the hardcoded routing table. The manifest maps statuses to workflows, phases, and agents.

### Module Discovery

If `_bmad/_config/manifest.yaml` exists, bmaduum reads installed modules. When SDET or TEA modules are detected, `test-automation` is injected into the lifecycle after `code-review`. Module info is shown in `--dry-run` output.

### bmad-help Fallback

When a story has a status the router doesn't recognize, bmaduum invokes `/bmad-help` via Claude CLI to determine the next workflow. This is depth-limited (max 3 recursive calls) to prevent infinite loops. Disable with `--no-bmad-help`.

---

## State File

The lifecycle executor persists execution state for error recovery.

**Location:** `.bmad-state.json` in the working directory.

**Lifecycle:**

1. **Saved on failure** - State is written when a workflow step fails
2. **Used on resume** - On re-run, execution continues from current status
3. **Cleared on success** - State file is deleted after successful completion
