# bmaduum

A CLI tool for automating [BMAD-METHOD](https://github.com/bmad-code-org/BMAD-METHOD) development workflows with Claude AI.

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/doc/go1.25)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**bmaduum** orchestrates Claude AI to automate development workflows—creating stories, implementing features, reviewing code, and managing git operations based on your project's sprint status.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Security Warning](#security-warning)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

```bash
# Install
go install ./cmd/bmaduum

# Run a story through its full lifecycle
bmaduum run 6-1-setup-project

# Preview what would run (dry-run)
bmaduum run 6-1-setup-project --dry-run

# Process entire epic
bmaduum epic 6

# Process all active epics
bmaduum all-epics
```

## Installation

### Prerequisites

- Go 1.25 or later
- [Claude CLI](https://github.com/anthropics/claude-code) installed and configured
- [just](https://github.com/casey/just) (optional, for running tasks)

### From Source

```bash
git clone https://github.com/yourusername/bmaduum.git
cd bmaduum
go install ./cmd/bmaduum
```

Or using just:

```bash
just install
```

### Build Only

```bash
just build
# Binary will be created as ./bmaduum
```

### Release Builds

Release builds with version information are available via GoReleaser:

```bash
just release-snapshot  # Local snapshot build
just release           # Full release (requires git tag)
```

## Usage

### Single Workflow Commands

```bash
# Create a story definition
bmaduum create-story <story-key> # eg 1-5

# Implement a story
bmaduum dev-story <story-key>

# Run code review
bmaduum code-review <story-key>

# Commit and push changes
bmaduum git-commit <story-key>
```

### Full Lifecycle

Run a story from its current status to completion:

```bash
bmaduum run <story-key>
```

This executes all remaining workflows based on the story's current status:

- `backlog` -> create-story -> dev-story -> code-review -> git-commit -> done
- `ready-for-dev` -> dev-story -> code-review -> git-commit -> done
- `in-progress` -> dev-story -> code-review -> git-commit -> done
- `review` -> code-review -> git-commit -> done
- `done` -> skipped (story already complete)

Status is automatically updated in `sprint-status.yaml` after each successful workflow.

Preview what workflows would run without executing them:

```bash
bmaduum run <story-key> --dry-run
```

### Epic Processing

Run the full lifecycle for all stories in one or more epics:

```bash
bmaduum epic <epic-id> [epic-id...]
```

This finds all stories matching the pattern `{epic-id}-{N}-*` (where N is numeric), sorts them by story number, and runs each to completion before moving to the next.

Examples:

```bash
# Single epic
bmaduum epic 6
# Runs 6-1-*, 6-2-*, 6-3-*, etc. each to completion in order

# Multiple epics
bmaduum epic 2 4 6
# Processes epics 2, 4, and 6 in sequence
```

The epic command stops on the first failure. Done stories are skipped.

#### Dry Run

Preview what workflows would run without executing them:

```bash
bmaduum epic 2 4 6 --dry-run
```

### All Epics

Run the full lifecycle for all active epics:

```bash
bmaduum all-epics
```

This command:
- Discovers all epics with non-completed stories
- Processes epics in numerical order
- Processes stories within each epic in order
- Stops on the first failure

Example:

```bash
bmaduum all-epics --dry-run  # Preview all workflows
bmaduum all-epics --auto-retry  # Run with automatic retry on rate limits
```

### Queue Processing

Run the full lifecycle for multiple stories in batch:

```bash
bmaduum queue <story-key> [story-key...]
```

Each story is run to completion before moving to the next. The queue stops on the first failure. Done stories are skipped.

Example:

```bash
bmaduum queue 6-5 6-6 6-7 6-8
```

Preview what workflows would run without executing them:

```bash
bmaduum queue 6-5 6-6 6-7 --dry-run
```

### Raw Prompts

Run an arbitrary prompt:

```bash
bmaduum raw "List all Go files in the project"
```

### Global Flags

| Flag           | Description                                           | Available On                     |
|----------------|-------------------------------------------------------|----------------------------------|
| `--dry-run`    | Preview workflows without executing them              | `run`, `queue`, `epic`, `all-epics` |
| `--auto-retry` | Automatically retry on failures (up to 10 times)      | `run`, `queue`, `epic`, `all-epics` |

### Help

```bash
bmaduum --help
bmaduum <command> --help
```

## Security Warning

This tool uses `--dangerously-skip-permissions` when invoking Claude CLI, which means:
- **No permission prompts** for file reads/writes
- **No confirmation** before command execution
- **Full automation** without user intervention

Only use in:
- Trusted repositories
- Isolated development environments
- CI/CD pipelines with restricted permissions

Do not use with:
- Untrusted codebases
- Production systems without safeguards

## How It Works

bmaduum spawns Claude CLI as a subprocess and manages execution through streaming JSON:

1. **Command** -> Parses story key and determines workflow from sprint-status.yaml
2. **Prompt** -> Expands Go template with story key (e.g., `"{{.StoryKey}}"` -> `"6-1-setup"`)
3. **Execute** -> Spawns `claude --dangerously-skip-permissions --output-format stream-json -p <prompt>`
4. **Parse** -> Reads streaming JSON events (text, tool use, tool results)
5. **Display** -> Formats output with progress indicators and styling
6. **Update** -> On success, updates status in sprint-status.yaml

```
+-------------+     +-------------+     +-------------+
|   Command   |---->|   Router    |---->|   Claude    |
|  run 6-1-*  |     |Status->Workflow|    | Subprocess  |
+-------------+     +-------------+     +------+------+
                                                |
                       +------------------------+
                       v
                +-------------+
                |    JSON     |
                |   Parser    |
                +------+------+
                       |
                       v
                +-------------+
                |   Styled    |
                |   Output    |
                +-------------+
```

## Configuration

### Configuration Precedence

Configuration is loaded in this priority order (highest wins):

1. **Environment variables** (`BMADUUM_*`)
2. **Config file** (specified by `BMADUUM_CONFIG_PATH`, or defaults to `./config/workflows.yaml`)
3. **Built-in defaults** (works out of the box without any config file)

### Config File

The config file is **optional**. If not found, sensible defaults are used.

To customize, create a file at one of these locations:
- `./config/workflows.yaml` (default, relative to working directory)
- A custom path specified via `BMADUUM_CONFIG_PATH` environment variable

Example config file:

```yaml
workflows:
  create-story:
    prompt_template: "Your custom prompt for {{.StoryKey}}"

  dev-story:
    prompt_template: "Your dev prompt for {{.StoryKey}}"

  code-review:
    prompt_template: "Your review prompt for {{.StoryKey}}"

  git-commit:
    prompt_template: "Your commit prompt for {{.StoryKey}}"

full_cycle:
  steps:
    - create-story
    - dev-story
    - code-review
    - git-commit

claude:
  output_format: stream-json
  binary_path: claude

output:
  truncate_lines: 20
  truncate_length: 60
```

#### What Each Section Controls

| Section | Purpose | Effect |
|---------|---------|--------|
| `workflows` | Defines prompts sent to Claude | Each workflow's `prompt_template` is expanded with the story key and passed to Claude CLI when that workflow runs. Customize to change how Claude behaves. |
| `full_cycle.steps` | Defines execution order | Controls which workflows run (and in what order) for `run`, `queue`, `epic`, and `all-epics` commands. |
| `claude.binary_path` | Claude CLI location | Path to the `claude` executable. Use this if Claude is not in your PATH or you want a specific version. |
| `claude.output_format` | Claude output format | Should be `stream-json` (required for parsing). Do not change unless you know what you are doing. |
| `output.truncate_lines` | Output display limit | Maximum number of lines to show per tool output. Additional lines are hidden with "... (N more lines)". |
| `output.truncate_length` | Line length limit | Maximum characters per line before truncation with "...". |

### Environment Variables

| Variable                     | Description                     | Default                   |
| ---------------------------- | ------------------------------- | ------------------------- |
| `BMADUUM_CONFIG_PATH`           | Path to custom config file      | `./config/workflows.yaml` |
| `BMADUUM_CLAUDE_PATH`           | Path to claude command/binary   | `claude`                  |
| `BMADUUM_CLAUDE_OUTPUT_FORMAT`  | Output format (stream-json)     | `stream-json`             |
| `BMADUUM_OUTPUT_TRUNCATE_LINES` | Max lines to display per event  | `20`                      |
| `BMADUUM_OUTPUT_TRUNCATE_LENGTH`| Max characters per output line  | `60`                      |

**Note:** `BMADUUM_CLAUDE_PATH` is a special shortcut for setting the claude command/binary path. For other config fields, use the `BMADUUM_` prefix with underscore-separated nested keys (e.g., `BMADUUM_OUTPUT_TRUNCATE_LINES` overrides `output.truncate_lines`).

**Example: Using a custom config file location**
```bash
export BMADUUM_CONFIG_PATH=~/my-configs/bmad-workflows.yaml
bmaduum run 6-1-setup
```

### Sprint Status File

The `run`, `queue`, `epic`, and `all-epics` commands read and update story status from:

```
_bmad-output/implementation-artifacts/sprint-status.yaml
```

Example format:

```yaml
development_status:
  6-1-setup-project: done
  6-2-add-feature: in-progress
  6-3-fix-bug: backlog
```

Valid status values:

| Status          | Description                       |
| --------------- | --------------------------------- |
| `backlog`       | Story not yet started             |
| `ready-for-dev` | Story ready for implementation    |
| `in-progress`   | Story currently being implemented |
| `review`        | Story in code review              |
| `done`          | Story complete                    |

### Prompt Templates

Workflow prompts use Go's `text/template` syntax. The available variable is:

- `{{.StoryKey}}` - The story identifier (e.g., `6-1-setup`)

Example: `"Work on story: {{.StoryKey}}"` expands to `"Work on story: 6-1-setup"`

### Important: Hardcoded Paths

The sprint status file path is **not configurable**. It must be located at:

```
_bmad-output/implementation-artifacts/sprint-status.yaml
```

Run `bmaduum` from your project root where this file exists.

## Troubleshooting

### "story not found" error
Ensure `_bmad-output/implementation-artifacts/sprint-status.yaml` exists and contains the story key under `development_status:`.

### Claude not found
If Claude CLI is not in your PATH, set the full path:
```bash
export BMADUUM_CLAUDE_PATH=/opt/homebrew/bin/claude
```

### Rate limiting
Use `--auto-retry` flag to automatically retry on rate limit errors with exponential backoff. The tool will detect rate limit errors from Claude's stderr and wait until the reset time before retrying.

### State persistence
If a lifecycle execution fails, state is saved to `.bmad-state.json` for potential resume functionality. The state is automatically cleared on successful completion.

## Development

### Prerequisites

- Go 1.25+
- [just](https://github.com/casey/just) command runner
- [golangci-lint](https://golangci-lint.run/) (for linting)
- [goreleaser](https://goreleaser.com/) (for release builds)

### Available Tasks

```bash
just              # Show all available tasks
just build        # Build the binary
just test         # Run all tests
just test-verbose # Run tests with verbose output
just test-coverage # Generate coverage report
just lint         # Run linter
just fmt          # Format code
just vet          # Run go vet
just check        # Run fmt, vet, and test
just clean        # Remove build artifacts
just release-snapshot  # Build release locally (snapshot)
just release           # Full release with GoReleaser
```

### Project Structure

```
bmaduum/
+-- cmd/bmaduum/     # Application entry point
+-- internal/
|   +-- claude/            # Claude client and JSON parser
|   +-- cli/               # Cobra CLI commands
|   +-- config/            # Configuration loading (Viper)
|   +-- lifecycle/         # Story lifecycle execution
|   +-- output/            # Terminal output formatting
|   +-- ratelimit/         # Rate limit detection
|   +-- router/            # Status-based workflow routing
|   +-- state/             # Execution state persistence
|   +-- status/            # Sprint status file reader/writer
|   +-- workflow/          # Workflow orchestration
+-- justfile               # Task runner configuration
+-- README.md
```

### Testing

Run tests:

```bash
just test
```

Run tests with coverage:

```bash
just test-coverage
# Open coverage.html in your browser
```

Test a specific package:

```bash
just test-pkg ./internal/claude
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
