# bmaduum

A CLI tool for automating [BMAD-METHOD](https://github.com/bmad-code-org/BMAD-METHOD) development workflows with Claude AI.

**bmaduum** orchestrates Claude AI to automate development workflows—creating stories, implementing features, reviewing code, and managing git operations based on your project's sprint status. It integrates with BMAD-METHOD v6 slash commands, letting the BMAD workflow engine handle agent personas, step-by-step execution, and progressive disclosure.

> **Warning:** This tool runs Claude CLI with `--dangerously-skip-permissions`, meaning Claude can read, write, and execute commands **without asking for confirmation**. Only use in trusted repositories and isolated environments.

## Installation

Requires [Claude CLI](https://github.com/anthropics/claude-code) installed and configured.

```bash
git clone https://github.com/ibro45/bmaduum.git
cd bmaduum
go install ./cmd/bmaduum
```

> **Note:** Installs to `~/go/bin`. Add to PATH: `export PATH="$HOME/go/bin:$PATH"`

## Usage

```bash
# Run a story through its full lifecycle
bmaduum story 6-1-setup-project

# Process multiple stories
bmaduum story 6-1 6-2 6-3

# Process an epic (all stories matching 6-*)
bmaduum epic 6

# Process all active epics
bmaduum epic all

# Preview without executing
bmaduum story --dry-run 6-1
bmaduum epic --dry-run all

# Run arbitrary prompt
bmaduum raw "List all Go files"
```

### Lifecycle

Stories progress through workflows based on their status in `sprint-status.yaml`:

| Status | Remaining Workflows |
|--------|---------------------|
| `backlog` | create-story -> dev-story -> code-review -> git-commit |
| `ready-for-dev` | dev-story -> code-review -> git-commit |
| `in-progress` | dev-story -> code-review -> git-commit |
| `review` | code-review -> git-commit |
| `done` | skipped |

If SDET or TEA modules are installed, `test-automation` is automatically added after `code-review`.

For unrecognized statuses, the optional bmad-help fallback invokes `/bmad-help` via Claude to determine the next workflow.

### Flags

| Flag | Commands | Description |
|------|----------|-------------|
| `--dry-run` | story, epic | Preview workflows without executing |
| `--auto-retry` | story, epic, workflow | Retry on rate limit errors |
| `--no-bmad-help` | story, epic | Disable bmad-help fallback for unknown statuses |

## Configuration

Configuration is optional. Defaults work out of the box with BMAD v6 slash commands.

```yaml
# config/workflows.yaml
use_slash_commands: true  # false for pre-v6 BMAD projects

workflows:
  dev-story:
    slash_command: "/dev-story {{.StoryKey}}"
    prompt_template: "/bmad-bmm-dev-story - Work on story: {{.StoryKey}}..."
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BMADUUM_CONFIG_PATH` | Path to configuration file | auto-discovered |
| `BMADUUM_CLAUDE_PATH` | Path to claude binary | `claude` |
| `BMADUUM_SPRINT_STATUS_PATH` | Path to sprint-status.yaml | auto-discovered |

### Sprint Status Path

The tool auto-discovers `sprint-status.yaml` in priority order:

1. `BMADUUM_SPRINT_STATUS_PATH` environment variable
2. `status_path` in config file
3. `_bmad-output/implementation-artifacts/sprint-status.yaml` (v6 path)
4. `sprint-status.yaml` (legacy path)

See [docs/CLI_REFERENCE.md](docs/CLI_REFERENCE.md) for full configuration options.

## Development

```bash
just build    # Build binary
just test     # Run tests
just lint     # Run linter
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for details.

## License

MIT - see [LICENSE](LICENSE)
