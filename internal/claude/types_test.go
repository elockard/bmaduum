package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEventFromStream_SystemInit(t *testing.T) {
	raw := &StreamEvent{
		Type:    "system",
		Subtype: "init",
	}

	event := NewEventFromStream(raw)

	assert.Equal(t, EventTypeSystem, event.Type)
	assert.Equal(t, "init", event.Subtype)
	assert.True(t, event.SessionStarted)
	assert.False(t, event.SessionComplete)
}

func TestNewEventFromStream_AssistantText(t *testing.T) {
	raw := &StreamEvent{
		Type: "assistant",
		Message: &MessageContent{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "Hello, I'm Claude!",
				},
			},
		},
	}

	event := NewEventFromStream(raw)

	assert.Equal(t, EventTypeAssistant, event.Type)
	assert.Equal(t, "Hello, I'm Claude!", event.Text)
	assert.True(t, event.IsText())
	assert.False(t, event.IsToolUse())
}

func TestNewEventFromStream_AssistantToolUse(t *testing.T) {
	raw := &StreamEvent{
		Type: "assistant",
		Message: &MessageContent{
			Content: []ContentBlock{
				{
					Type: "tool_use",
					Name: "Bash",
					Input: &ToolInput{
						Command:     "ls -la",
						Description: "List files",
					},
				},
			},
		},
	}

	event := NewEventFromStream(raw)

	assert.Equal(t, EventTypeAssistant, event.Type)
	assert.Equal(t, "Bash", event.ToolName)
	assert.Equal(t, "ls -la", event.ToolCommand)
	assert.Equal(t, "List files", event.ToolDescription)
	assert.True(t, event.IsToolUse())
	assert.False(t, event.IsText())
}

func TestNewEventFromStream_ToolResult(t *testing.T) {
	raw := &StreamEvent{
		Type: "user",
		ToolUseResult: &ToolResult{
			Stdout: "file1.go\nfile2.go",
			Stderr: "",
		},
	}

	event := NewEventFromStream(raw)

	assert.Equal(t, EventTypeUser, event.Type)
	assert.Equal(t, "file1.go\nfile2.go", event.ToolStdout)
	assert.True(t, event.IsToolResult())
}

func TestNewEventFromStream_Result(t *testing.T) {
	raw := &StreamEvent{
		Type: "result",
	}

	event := NewEventFromStream(raw)

	assert.Equal(t, EventTypeResult, event.Type)
	assert.True(t, event.SessionComplete)
}

func TestEvent_IsText(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name:     "text event",
			event:    Event{Type: EventTypeAssistant, Text: "hello"},
			expected: true,
		},
		{
			name:     "empty text",
			event:    Event{Type: EventTypeAssistant, Text: ""},
			expected: false,
		},
		{
			name:     "wrong type",
			event:    Event{Type: EventTypeSystem, Text: "hello"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.IsText())
		})
	}
}

func TestEvent_IsToolUse(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name:     "tool use event",
			event:    Event{Type: EventTypeAssistant, ToolName: "Bash"},
			expected: true,
		},
		{
			name:     "empty tool name",
			event:    Event{Type: EventTypeAssistant, ToolName: ""},
			expected: false,
		},
		{
			name:     "wrong type",
			event:    Event{Type: EventTypeSystem, ToolName: "Bash"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.IsToolUse())
		})
	}
}

func TestEvent_IsToolResult(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name:     "tool result with stdout",
			event:    Event{Type: EventTypeUser, ToolStdout: "output"},
			expected: true,
		},
		{
			name:     "tool result with stderr",
			event:    Event{Type: EventTypeUser, ToolStderr: "error"},
			expected: true,
		},
		{
			name:     "empty result",
			event:    Event{Type: EventTypeUser},
			expected: false,
		},
		{
			name:     "wrong type",
			event:    Event{Type: EventTypeAssistant, ToolStdout: "output"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.IsToolResult())
		})
	}
}
