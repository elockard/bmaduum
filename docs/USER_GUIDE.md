# User Guide

A practical guide to using `bmaduum` for automating development workflows.

## Overview

`bmaduum` is a CLI tool that orchestrates Claude AI to automate repetitive development tasks. It handles:

- Creating story definitions from story keys
- Implementing features based on story requirements
- Running code reviews
- Committing and pushing changes
- Processing multiple stories in batch
- Managing entire epics

## Quick Start

### Prerequisites

1. **Go 1.21+** installed
2. **Claude CLI** installed and configured ([installation guide](https://github.com/anthropics/claude-code))
3. **just** command runner (optional but recommended)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/bmaduum.git
cd bmaduum

# Build the binary
just build
# OR
go build -o bmaduum ./cmd/bmaduum

# (Optional) Install globally
go install ./cmd/bmaduum
```

### Verify Installation

```bash
bmaduum --help
```

## Basic Usage

### Working with Individual Stories

#### 1. Create a Story

Generate a story definition from a story key:

```bash
bmaduum create-story PROJ-123
```

This runs Claude with a prompt to create the story definition based on your configured template.

#### 2. Implement the Story

Run the development workflow:

```bash
bmaduum dev-story PROJ-123
```

Claude will:

- Read the story requirements
- Implement the feature
- Run tests after each change

#### 3. Review the Code

Run code review on your changes:

```bash
bmaduum code-review PROJ-123
```

Claude reviews the code and automatically fixes issues.

#### 4. Commit and Push

Create a commit and push to remote:

```bash
bmaduum git-commit PROJ-123
```

### Status-Based Automation

Instead of manually running each step, let `bmaduum` determine what to do:

```bash
bmaduum run PROJ-123
```

The `run` command now executes the **complete lifecycle** from the story's current status all the way to `done`. It auto-updates the status after each successful step and stops at completion or on the first failure.

| Story Status    | Remaining Lifecycle                                            |
| --------------- | -------------------------------------------------------------- |
| `backlog`       | create-story -> dev-story -> code-review -> git-commit -> done |
| `ready-for-dev` | dev-story -> code-review -> git-commit -> done                 |
| `in-progress`   | dev-story -> code-review -> git-commit -> done                 |
| `review`        | code-review -> git-commit -> done                              |
| `done`          | No action (story already complete)                             |

### Full Lifecycle Execution

The `run` command processes stories through their entire remaining lifecycle automatically:

```bash
bmaduum run PROJ-123
```

**Example output:**

```
[1/4] create-story
  ... workflow output ...

[2/4] dev-story
  ... workflow output ...

[3/4] code-review
  ... workflow output ...

[4/4] git-commit
  ... workflow output ...

Story PROJ-123 completed successfully
```

**How it works:**

1. Reads the story's current status from `sprint-status.yaml`
2. Determines remaining lifecycle steps based on status
3. Executes each workflow in sequence
4. Updates status in `sprint-status.yaml` after each successful step
5. Stops at `done` or on first failure

### Dry Run Mode

Preview what workflows will run without executing them:

```bash
bmaduum run --dry-run PROJ-123
```

**Example output:**

```
Dry run for story PROJ-123:
  1. create-story -> ready-for-dev
  2. dev-story -> review
  3. code-review -> done
  4. git-commit -> done
```

Dry run is available for all lifecycle commands:

```bash
bmaduum run --dry-run PROJ-123           # Single story
bmaduum queue --dry-run PROJ-123 PROJ-124  # Multiple stories
bmaduum epic --dry-run 05                # All stories in epic
```

### Error Recovery

When a workflow fails, the tool saves execution state so you can resume from the point of failure:

**State file:** `.bmad-state.json`

```json
{
	"story_key": "PROJ-123",
	"step_index": 2,
	"total_steps": 4,
	"start_status": "backlog"
}
```

**Resuming after failure:**

```bash
# Workflow fails at step 2 (dev-story)
bmaduum run PROJ-123
# Error: workflow failed: dev-story returned exit code 1

# Fix the issue, then re-run
bmaduum run PROJ-123
# Continues from current status (no work is lost)
```

The state file is automatically cleared on successful completion.

### Batch Processing

Process multiple stories at once:

```bash
bmaduum queue PROJ-123 PROJ-124 PROJ-125
```

The queue:

- Runs each story through its **full lifecycle** to completion
- Auto-updates status after each successful workflow step
- Skips completed stories
- Stops on first failure
- Shows a summary at the end

Use `--dry-run` to preview the full execution plan:

```bash
bmaduum queue --dry-run PROJ-123 PROJ-124 PROJ-125
```

### Processing Epics

Run all stories in one or more epics through their full lifecycle:

```bash
# Single epic
bmaduum epic 05

# Multiple epics
bmaduum epic 02 04 06
```

This finds all stories matching the pattern `{epic-id}-{N}-*` (e.g., `05-01-auth`, `05-02-dashboard`), sorts them by story number, and runs each through its complete lifecycle to completion.

Use `--dry-run` to preview:

```bash
bmaduum epic --dry-run 02 04 06
```

### Ad-Hoc Prompts

Run any prompt directly:

```bash
bmaduum raw "List all TODO comments in the codebase"
bmaduum raw "What tests are missing coverage?"
```

## Configuration

### Config File Location

By default, configuration is loaded from `config/workflows.yaml`.

Override with environment variable:

```bash
export BMADUUM_CONFIG_PATH=/path/to/custom/config.yaml
bmaduum run PROJ-123
```

### Customizing Workflows

Edit `config/workflows.yaml` to customize workflow prompts:

```yaml
workflows:
  create-story:
    prompt_template: |
      Create a detailed story definition for {{.StoryKey}}.
      Include acceptance criteria and technical requirements.
      Do not ask clarifying questions.

  dev-story:
    prompt_template: |
      Implement story {{.StoryKey}}.
      Follow existing code patterns.
      Run tests after each change.
      Do not ask questions - use best judgment.

  code-review:
    prompt_template: |
      Review changes for story {{.StoryKey}}.
      Check for:
      - Code quality issues
      - Missing tests
      - Security vulnerabilities
      Auto-fix all issues immediately.

  git-commit:
    prompt_template: |
      Commit changes for {{.StoryKey}}.
      Use conventional commit format.
      Push to current branch.
```

### Template Variables

| Variable        | Description                         |
| --------------- | ----------------------------------- |
| `{{.StoryKey}}` | The story key passed to the command |

### Output Settings

Control how much output is displayed:

```yaml
output:
  truncate_lines: 20 # Max lines for tool output
  truncate_length: 60 # Max chars for command headers
```

### Claude Settings

Customize Claude execution:

```yaml
claude:
  binary_path: claude # Path to Claude binary
  output_format: stream-json # Output format (don't change)
```

Or use environment variables:

```bash
export BMADUUM_CLAUDE_PATH=/usr/local/bin/claude
```

## Sprint Status File

### File Location

The tool reads story status from:

```
_bmad-output/implementation-artifacts/sprint-status.yaml
```

### File Format

```yaml
development_status:
  PROJ-123: ready-for-dev
  PROJ-124: in-progress
  PROJ-125: review
  PROJ-126: done
  05-01-auth: backlog
  05-02-dashboard: ready-for-dev
```

### Valid Status Values

| Status          | Meaning                                 |
| --------------- | --------------------------------------- |
| `backlog`       | Not started, needs story creation       |
| `ready-for-dev` | Story created, ready for implementation |
| `in-progress`   | Currently being implemented             |
| `review`        | Implementation done, needs review       |
| `done`          | Complete                                |

## Workflow Patterns

### Pattern 1: Sequential Development

Run each step manually for full control:

```bash
bmaduum create-story PROJ-123
# Review the story definition
bmaduum dev-story PROJ-123
# Test the implementation manually
bmaduum code-review PROJ-123
# Verify fixes
bmaduum git-commit PROJ-123
```

### Pattern 2: Status-Driven

Let the tool figure out what to do:

```bash
# Run whatever step is needed next
bmaduum run PROJ-123

# Run again after updating status
bmaduum run PROJ-123
```

### Pattern 3: Batch Sprint

Process an entire sprint's stories:

```bash
bmaduum queue SPRINT-1 SPRINT-2 SPRINT-3 SPRINT-4 SPRINT-5
```

### Pattern 4: Epic Processing

Process all stories in an epic:

```bash
bmaduum epic 05
```

### Pattern 5: Investigation

Use raw prompts for ad-hoc tasks:

```bash
# Understand the codebase
bmaduum raw "Explain the authentication flow"

# Find issues
bmaduum raw "What tests have the most failures?"

# Generate reports
bmaduum raw "Create a summary of recent changes"
```

## Understanding Output

### Tool Invocations

When Claude uses tools, you'll see formatted output:

```
┌─ Bash ─────────────────────────────────────────────────────────
│  List project files
│  $ ls -la
└────────────────────────────────────────────────────────────────

total 48
drwxr-xr-x  12 user  staff   384 Jan  8 10:00 .
drwxr-xr-x   5 user  staff   160 Jan  8 09:00 ..
...
```

### Progress Indicators

| Symbol | Meaning     |
| ------ | ----------- |
| ●      | In progress |
| ✓      | Success     |
| ✗      | Failure     |
| ○      | Skipped     |

### Queue Summary

After processing multiple stories:

```
Summary:
  PROJ-123  ✓  1m 23s
  PROJ-124  ✓  2m 45s
  PROJ-125  ○  skipped (done)
  PROJ-126  ✗  failed at dev-story (45s)
```

## Error Handling

### Exit Codes

| Code | Meaning                             |
| ---- | ----------------------------------- |
| 0    | Success                             |
| 1    | General error                       |
| N    | Claude's exit code (passed through) |

### Resume Capability

When a lifecycle execution fails, the tool persists state to `.bmad-state.json`. This enables resume from the point of failure:

1. **Failure occurs** - State is saved with current story and step index
2. **Fix the issue** - Address whatever caused the failure
3. **Re-run the command** - Execution continues from current status

The state file is automatically deleted after successful lifecycle completion.

### State File Location

```
.bmad-state.json   # In working directory (hidden file)
```

You can safely delete this file to force a fresh start from the story's current status.

### Common Issues

**Claude not found:**

```
Error: failed to start claude: exec: "claude": executable file not found in $PATH
```

Solution: Install Claude CLI or set `BMADUUM_CLAUDE_PATH`.

**Config not found:**

```
Error: error loading config: open config/workflows.yaml: no such file or directory
```

Solution: Create the config file or set `BMADUUM_CONFIG_PATH`.

**Story not in status file:**

```
Error: story PROJ-999 not found in sprint status
```

Solution: Add the story to `sprint-status.yaml`.

**Unknown status:**

```
Error: unknown status value: invalid-status
```

Solution: Use a valid status: `backlog`, `ready-for-dev`, `in-progress`, `review`, or `done`.

## Tips and Best Practices

### 1. Start Small

Begin with single story commands to understand the workflow:

```bash
bmaduum create-story TEST-001
```

### 2. Review Output

Always review Claude's output before moving to the next step. The streaming output shows exactly what Claude is doing.

### 3. Use Descriptive Story Keys

Story keys appear in commits and prompts. Use meaningful identifiers:

- Good: `AUTH-001`, `FEAT-user-profile`, `BUG-login-fix`
- Bad: `x`, `test`, `123`

### 4. Customize Prompts for Your Project

Edit the prompt templates to match your project's conventions, coding standards, and requirements.

### 5. Keep Status File Updated

The status file is the source of truth for automation. Keep it current as stories progress.

### 6. Handle Failures Gracefully

When a queue stops on failure:

1. Review the error
2. Fix the issue manually or adjust the story
3. Update the status file
4. Re-run the queue (completed stories will be skipped)

### 7. Use Raw for Exploration

Before starting a story, use raw prompts to understand the codebase:

```bash
bmaduum raw "What files would I need to change to add user authentication?"
```
