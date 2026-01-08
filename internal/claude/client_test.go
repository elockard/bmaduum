package claude

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockExecutor_Execute(t *testing.T) {
	events := []Event{
		{Type: EventTypeSystem, SessionStarted: true},
		{Type: EventTypeAssistant, Text: "Hello!"},
		{Type: EventTypeResult, SessionComplete: true},
	}

	mock := &MockExecutor{Events: events}

	ctx := context.Background()
	ch, err := mock.Execute(ctx, "test prompt")

	require.NoError(t, err)

	// Collect all events
	var collected []Event
	for event := range ch {
		collected = append(collected, event)
	}

	assert.Equal(t, events, collected)
	assert.Equal(t, []string{"test prompt"}, mock.RecordedPrompts)
}

func TestMockExecutor_Execute_WithError(t *testing.T) {
	mock := &MockExecutor{
		Error: assert.AnError,
	}

	ctx := context.Background()
	_, err := mock.Execute(ctx, "test prompt")

	assert.Error(t, err)
}

func TestMockExecutor_ExecuteWithResult(t *testing.T) {
	events := []Event{
		{Type: EventTypeSystem, SessionStarted: true},
		{Type: EventTypeAssistant, Text: "Hello!"},
		{Type: EventTypeResult, SessionComplete: true},
	}

	mock := &MockExecutor{
		Events:   events,
		ExitCode: 0,
	}

	var receivedEvents []Event
	handler := func(event Event) {
		receivedEvents = append(receivedEvents, event)
	}

	ctx := context.Background()
	exitCode, err := mock.ExecuteWithResult(ctx, "test prompt", handler)

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, events, receivedEvents)
}

func TestMockExecutor_ExecuteWithResult_NonZeroExit(t *testing.T) {
	mock := &MockExecutor{
		ExitCode: 1,
	}

	ctx := context.Background()
	exitCode, err := mock.ExecuteWithResult(ctx, "test prompt", nil)

	require.NoError(t, err)
	assert.Equal(t, 1, exitCode)
}

func TestMockExecutor_Execute_ContextCancellation(t *testing.T) {
	events := []Event{
		{Type: EventTypeSystem, SessionStarted: true},
		{Type: EventTypeAssistant, Text: "Hello!"},
		{Type: EventTypeResult, SessionComplete: true},
	}

	mock := &MockExecutor{Events: events}

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := mock.Execute(ctx, "test prompt")

	require.NoError(t, err)

	// Read first event
	<-ch

	// Cancel context
	cancel()

	// Should eventually close (may receive more events before closing)
	for range ch {
		// Drain channel
	}
}

func TestNewExecutor(t *testing.T) {
	// Test default config
	exec := NewExecutor(ExecutorConfig{})
	assert.NotNil(t, exec)
	assert.Equal(t, "claude", exec.config.BinaryPath)
	assert.Equal(t, "stream-json", exec.config.OutputFormat)

	// Test custom config
	exec = NewExecutor(ExecutorConfig{
		BinaryPath:   "/custom/claude",
		OutputFormat: "json",
	})
	assert.Equal(t, "/custom/claude", exec.config.BinaryPath)
	assert.Equal(t, "json", exec.config.OutputFormat)
}
