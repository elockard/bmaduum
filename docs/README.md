# bmaduum Documentation

Comprehensive documentation for the `bmaduum` CLI tool.

## Documentation Index

### For Users

| Document                           | Description                                   |
| ---------------------------------- | --------------------------------------------- |
| [User Guide](USER_GUIDE.md)        | Getting started, installation, usage patterns |
| [CLI Reference](CLI_REFERENCE.md)  | Complete command reference with examples      |
| [CLI Cookbook](examples/README.md) | Recipe-style CLI examples                     |

### For Developers

| Document                             | Description                            |
| ------------------------------------ | -------------------------------------- |
| [Architecture](ARCHITECTURE.md)      | System design, diagrams, data flow     |
| [Package Documentation](PACKAGES.md) | API reference for all packages         |
| [Development Guide](DEVELOPMENT.md)  | Setup, testing, extending the codebase |

### Quick Links

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands Overview](#commands-overview)
- [Configuration](#configuration)

## Installation

```bash
# Clone and build
git clone https://github.com/yourusername/bmaduum.git
cd bmaduum
just build

# Or install globally
go install ./cmd/bmaduum
```

## Quick Start

```bash
# Process a story through its full lifecycle to done
# (automatically runs: create-story -> dev-story -> code-review -> git-commit)
bmaduum run PROJ-123

# Preview what workflows would run without executing
bmaduum run --dry-run PROJ-123

# Process multiple stories through their lifecycles
bmaduum queue PROJ-123 PROJ-124 PROJ-125

# Process one or more epics
bmaduum epic 05
bmaduum epic 02 04 06

# Run an arbitrary prompt
bmaduum raw "What files need tests?"
```

## Commands Overview

| Command        | Purpose                                        |
| -------------- | ---------------------------------------------- |
| `create-story` | Create story definition                        |
| `dev-story`    | Implement a story                              |
| `code-review`  | Review code changes                            |
| `git-commit`   | Commit and push                                |
| `run`          | Execute full lifecycle to done (with resume)   |
| `queue`        | Batch process stories through lifecycles       |
| `epic`         | Process all stories in epic through lifecycles |
| `raw`          | Execute arbitrary prompt                       |

All lifecycle commands (`run`, `queue`, `epic`) support `--dry-run` to preview execution.

See [CLI Reference](CLI_REFERENCE.md) for complete details.

## Configuration

Default configuration file: `config/workflows.yaml`

```yaml
workflows:
  create-story:
    prompt_template: "Create story: {{.StoryKey}}"
  dev-story:
    prompt_template: "Implement story: {{.StoryKey}}"
  code-review:
    prompt_template: "Review story: {{.StoryKey}}"
  git-commit:
    prompt_template: "Commit changes for: {{.StoryKey}}"
```

See [User Guide](USER_GUIDE.md#configuration) for complete configuration options.

## Architecture Overview

```
cmd/bmaduum/main.go
         │
         ▼
    internal/cli (Cobra commands)
         │
         ├──► internal/lifecycle (lifecycle orchestration)
         │         │
         │         └──► internal/workflow (single workflow execution)
         │
         ├──► internal/state (execution state persistence)
         │
         ├──► internal/router (status → workflow routing)
         │
         └──► internal/config (configuration)
```

See [Architecture](ARCHITECTURE.md) for detailed diagrams and explanations.

## Contributing

See [Development Guide](DEVELOPMENT.md) for:

- Development setup
- Testing practices
- Adding new commands
- Code style guidelines
