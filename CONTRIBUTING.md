# Contributing to bmad-automate

Thank you for your interest in contributing to bmad-automate! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/bmad-automate.git`
3. Create a branch for your changes: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `just check`
6. Commit your changes with a descriptive message
7. Push to your fork and submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- [just](https://github.com/casey/just) command runner
- [golangci-lint](https://golangci-lint.run/) for linting

### Building

```bash
just build
```

### Running Tests

```bash
# Run all tests
just test

# Run with verbose output
just test-verbose

# Run with coverage
just test-coverage

# Test a specific package
just test-pkg ./internal/claude
```

### Code Quality

Before submitting a PR, please run:

```bash
just check
```

This runs:

- `go fmt` - Format code
- `go vet` - Static analysis
- `go test` - Run tests

For more thorough linting:

```bash
just lint
```

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (run `just fmt`)
- Write meaningful commit messages
- Add tests for new functionality
- Update documentation as needed

## Project Structure

```text
bmad-automate/
├── cmd/bmad-automate/     # Application entry point
├── config/                # Default configuration files
├── internal/
│   ├── cli/               # Cobra CLI commands
│   ├── claude/            # Claude client and JSON parser
│   ├── config/            # Configuration loading (Viper)
│   ├── output/            # Terminal output (Lipgloss)
│   └── workflow/          # Workflow orchestration
```

### Package Guidelines

- **cli**: Thin command handlers that delegate to other packages
- **claude**: Claude CLI interaction, streaming JSON parsing
- **config**: Configuration loading and validation
- **output**: All terminal output and styling
- **workflow**: Business logic for workflow execution

## Testing Guidelines

- Write unit tests for new functionality
- Use table-driven tests where appropriate
- Use `testify` for assertions and mocking
- Aim for good coverage on business logic packages
- Test files should be named `*_test.go`

### Mocking

The codebase uses interfaces for testability:

- `claude.Executor` - Mock Claude execution
- `output.Printer` - Capture output for testing

Example:

```go
func TestMyFeature(t *testing.T) {
    mock := &claude.MockExecutor{
        Events: []claude.Event{...},
        ExitCode: 0,
    }
    // Use mock in tests
}
```

## Pull Request Process

1. Ensure all tests pass (`just check`)
2. Update documentation if needed
3. Add a clear description of your changes
4. Reference any related issues
5. Request review from maintainers

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Any relevant logs or error messages

## Feature Requests

Feature requests are welcome! Please:

- Check existing issues first
- Describe the use case
- Explain why this would be useful
- Consider if you'd like to implement it

## Questions?

Feel free to open an issue for questions or discussions about the project.
