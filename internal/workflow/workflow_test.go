package workflow

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bmaduum/internal/claude"
	"bmaduum/internal/config"
	"bmaduum/internal/output"
)

func setupTestRunner() (*Runner, *claude.MockExecutor, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)
	cfg := config.DefaultConfig()
	mockExecutor := &claude.MockExecutor{
		Events: []claude.Event{
			{Type: claude.EventTypeSystem, SessionStarted: true},
			{Type: claude.EventTypeAssistant, Text: "Working on it..."},
			{Type: claude.EventTypeResult, SessionComplete: true},
		},
		ExitCode: 0,
	}
	runner := NewRunner(mockExecutor, printer, cfg)
	return runner, mockExecutor, buf
}

func TestNewRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	printer := output.NewPrinter()
	executor := &claude.MockExecutor{}

	runner := NewRunner(executor, printer, cfg)

	assert.NotNil(t, runner)
}

func TestRunner_RunSingle(t *testing.T) {
	runner, mockExecutor, _ := setupTestRunner()

	ctx := context.Background()
	exitCode := runner.RunSingle(ctx, "create-story", "test-123")

	assert.Equal(t, 0, exitCode)
	require.Len(t, mockExecutor.RecordedPrompts, 1)
	assert.Contains(t, mockExecutor.RecordedPrompts[0], "test-123")
}

func TestRunner_RunSingle_UnknownWorkflow(t *testing.T) {
	runner, _, _ := setupTestRunner()

	ctx := context.Background()
	exitCode := runner.RunSingle(ctx, "unknown-workflow", "test-123")

	assert.Equal(t, 1, exitCode)
}

func TestRunner_RunRaw(t *testing.T) {
	runner, mockExecutor, _ := setupTestRunner()

	ctx := context.Background()
	exitCode := runner.RunRaw(ctx, "custom prompt")

	assert.Equal(t, 0, exitCode)
	require.Len(t, mockExecutor.RecordedPrompts, 1)
	assert.Equal(t, "custom prompt", mockExecutor.RecordedPrompts[0])
}

func TestRunner_HandleEvent(t *testing.T) {
	runner, _, buf := setupTestRunner()

	// Test session start
	runner.handleEvent(claude.Event{Type: claude.EventTypeSystem, SessionStarted: true})
	assert.Contains(t, buf.String(), "Session started")

	buf.Reset()

	// Test text
	runner.handleEvent(claude.Event{Type: claude.EventTypeAssistant, Text: "Hello!"})
	assert.Contains(t, buf.String(), "Hello!")

	buf.Reset()

	// Test tool use with result - tool uses are buffered until their results arrive
	// to enable printing tool+result pairs together (matching Claude Code style)
	runner.handleEvent(claude.Event{
		Type:            claude.EventTypeAssistant,
		ToolID:          "tool-123",
		ToolName:        "Bash",
		ToolCommand:     "ls",
		ToolDescription: "List files",
	})
	// Tool use is buffered, not printed yet
	assert.NotContains(t, buf.String(), "Bash")

	// When result arrives, both tool use and result are printed together
	runner.handleEvent(claude.Event{
		Type:          claude.EventTypeUser,
		ToolUseID:     "tool-123",
		ToolStdout:    "file1.go",
		HasToolResult: true,
	})
	assert.Contains(t, buf.String(), "Bash")
	assert.Contains(t, buf.String(), "file1.go")
}

func TestRunner_HandleEvent_ToolUseFlushedOnText(t *testing.T) {
	runner, _, buf := setupTestRunner()

	// Tool use is buffered
	runner.handleEvent(claude.Event{
		Type:            claude.EventTypeAssistant,
		ToolID:          "tool-123",
		ToolName:        "Bash",
		ToolCommand:     "ls",
		ToolDescription: "List files",
	})
	assert.NotContains(t, buf.String(), "Bash")

	// When text arrives, buffered tools are flushed first
	runner.handleEvent(claude.Event{Type: claude.EventTypeAssistant, Text: "Done!"})
	assert.Contains(t, buf.String(), "Bash")
	assert.Contains(t, buf.String(), "Done!")
}
