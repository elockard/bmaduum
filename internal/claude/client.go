package claude

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Executor runs Claude CLI and returns streaming events.
type Executor interface {
	// Execute runs Claude with the given prompt and returns a channel of events.
	// The channel is closed when Claude exits.
	// Returns an error if Claude fails to start.
	// Note: This method is fire-and-forget; the exit status is not available.
	// Use ExecuteWithResult if you need the exit code.
	Execute(ctx context.Context, prompt string) (<-chan Event, error)

	// ExecuteWithResult runs Claude and waits for completion.
	// Returns the exit code and any error.
	ExecuteWithResult(ctx context.Context, prompt string, handler EventHandler) (int, error)
}

// EventHandler is called for each event received from Claude.
type EventHandler func(event Event)

// ExecutorConfig contains configuration for the Claude executor.
type ExecutorConfig struct {
	// BinaryPath is the path to the Claude binary.
	// Defaults to "claude" (found in PATH).
	BinaryPath string

	// OutputFormat is the output format flag.
	// Defaults to "stream-json".
	OutputFormat string

	// Parser is the JSON parser to use.
	// If nil, a DefaultParser is created.
	Parser Parser

	// StderrHandler is called for each line of stderr output.
	// If nil, stderr is ignored.
	StderrHandler func(line string)
}

// DefaultExecutor implements Executor using os/exec.
type DefaultExecutor struct {
	config ExecutorConfig
	parser Parser
}

// NewExecutor creates a new DefaultExecutor with the given configuration.
func NewExecutor(config ExecutorConfig) *DefaultExecutor {
	if config.BinaryPath == "" {
		config.BinaryPath = "claude"
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "stream-json"
	}

	parser := config.Parser
	if parser == nil {
		parser = NewParser()
	}

	return &DefaultExecutor{
		config: config,
		parser: parser,
	}
}

// Execute runs Claude with the given prompt and returns a channel of events.
// Note: This method is fire-and-forget; the exit status is not available.
// Use ExecuteWithResult if you need the exit code.
func (e *DefaultExecutor) Execute(ctx context.Context, prompt string) (<-chan Event, error) {
	cmd := exec.CommandContext(ctx, e.config.BinaryPath,
		"--dangerously-skip-permissions",
		"-p", prompt,
		"--output-format", e.config.OutputFormat,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	// Handle stderr in background
	go e.handleStderr(stderr)

	// Parse stdout and return events channel
	events := e.parser.Parse(stdout)

	// Wait for command completion in background.
	// Note: Exit status is intentionally not propagated; use ExecuteWithResult if needed.
	go func() {
		_ = cmd.Wait() //nolint:errcheck // Exit status intentionally ignored; use ExecuteWithResult if needed
	}()

	return events, nil
}

// ExecuteWithResult runs Claude and waits for completion.
func (e *DefaultExecutor) ExecuteWithResult(ctx context.Context, prompt string, handler EventHandler) (int, error) {
	cmd := exec.CommandContext(ctx, e.config.BinaryPath,
		"--dangerously-skip-permissions",
		"-p", prompt,
		"--output-format", e.config.OutputFormat,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return 1, fmt.Errorf("failed to start claude: %w", err)
	}

	// Handle stderr in background
	go e.handleStderr(stderr)

	// Process events
	events := e.parser.Parse(stdout)
	for event := range events {
		if handler != nil {
			handler(event)
		}
	}

	// Wait for command completion
	err = cmd.Wait()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return 1, err
		}
	}

	return exitCode, nil
}

func (e *DefaultExecutor) handleStderr(stderr io.ReadCloser) {
	if e.config.StderrHandler == nil {
		_, _ = io.Copy(io.Discard, stderr) //nolint:errcheck // Intentionally discarding stderr
		return
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		e.config.StderrHandler(scanner.Text())
	}
}

// MockExecutor implements Executor for testing.
type MockExecutor struct {
	// Events to return from Execute.
	Events []Event

	// Error to return from Execute.
	Error error

	// ExitCode to return from ExecuteWithResult.
	ExitCode int

	// RecordedPrompts stores prompts passed to Execute.
	RecordedPrompts []string
}

// Execute returns the pre-configured events.
func (m *MockExecutor) Execute(ctx context.Context, prompt string) (<-chan Event, error) {
	m.RecordedPrompts = append(m.RecordedPrompts, prompt)

	if m.Error != nil {
		return nil, m.Error
	}

	events := make(chan Event)
	go func() {
		defer close(events)
		for _, event := range m.Events {
			select {
			case <-ctx.Done():
				return
			case events <- event:
			}
		}
	}()

	return events, nil
}

// ExecuteWithResult returns the pre-configured exit code.
func (m *MockExecutor) ExecuteWithResult(ctx context.Context, prompt string, handler EventHandler) (int, error) {
	m.RecordedPrompts = append(m.RecordedPrompts, prompt)

	if m.Error != nil {
		return 1, m.Error
	}

	for _, event := range m.Events {
		if handler != nil {
			handler(event)
		}
	}

	return m.ExitCode, nil
}
