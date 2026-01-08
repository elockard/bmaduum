# bmad-automate

A CLI tool for automating development workflows with Claude AI.

## Overview

`bmad-automate` orchestrates Claude to run development workflows including story creation, implementation, code review, and git operations. It's designed to automate repetitive development tasks by delegating them to Claude with predefined prompts.

## Features

- **Workflow Automation** - Run predefined workflows (create-story, dev-story, code-review, git-commit)
- **Full Cycle Execution** - Execute all workflow steps in sequence with a single command
- **Queue Processing** - Process multiple stories in batch
- **Configurable Prompts** - Customize workflow prompts via YAML configuration
- **Streaming Output** - Real-time feedback from Claude's execution
- **Styled Terminal Output** - Clean, readable output with progress indicators

## Installation

### Prerequisites

- Go 1.21 or later
- [Claude CLI](https://github.com/anthropics/claude-code) installed and configured
- [just](https://github.com/casey/just) (optional, for running tasks)

### From Source

```bash
git clone https://github.com/yourusername/bmad-automate.git
cd bmad-automate
go install ./cmd/bmad-automate
```

Or using just:

```bash
just install
```

### Build Only

```bash
just build
# Binary will be created as ./bmad-automate
```

## Usage

### Single Workflow Commands

```bash
# Create a story definition
bmad-automate create-story <story-key>

# Implement a story
bmad-automate dev-story <story-key>

# Run code review
bmad-automate code-review <story-key>

# Commit and push changes
bmad-automate git-commit <story-key>
```

### Full Cycle

Run all workflow steps in sequence:

```bash
bmad-automate run <story-key>
```

This executes: `create-story` → `dev-story` → `code-review` → `git-commit`

### Queue Processing

Process multiple stories in batch:

```bash
bmad-automate queue story-1 story-2 story-3
```

The queue stops on the first failure.

### Raw Prompts

Run an arbitrary prompt:

```bash
bmad-automate raw "List all Go files in the project"
```

### Help

```bash
bmad-automate --help
bmad-automate <command> --help
```

## Configuration

### Config File

Create a `config/workflows.yaml` file to customize workflow prompts:

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

### Environment Variables

| Variable           | Description                | Default                   |
| ------------------ | -------------------------- | ------------------------- |
| `BMAD_CONFIG_PATH` | Path to custom config file | `./config/workflows.yaml` |
| `BMAD_CLAUDE_PATH` | Path to Claude binary      | `claude`                  |

## Development

### Prerequisites

- Go 1.21+
- [just](https://github.com/casey/just) command runner
- [golangci-lint](https://golangci-lint.run/) (for linting)

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
```

### Project Structure

```
bmad-automate/
├── cmd/bmad-automate/     # Application entry point
├── config/                # Default configuration
├── internal/
│   ├── cli/               # Cobra CLI commands
│   ├── claude/            # Claude client and parser
│   ├── config/            # Configuration loading
│   ├── output/            # Terminal output formatting
│   └── workflow/          # Workflow orchestration
├── justfile               # Task runner configuration
└── README.md
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
