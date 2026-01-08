package claude

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultParser_Parse(t *testing.T) {
	input := `{"type":"system","subtype":"init"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]}}
{"type":"result"}`

	parser := NewParser()
	reader := strings.NewReader(input)

	events := parser.Parse(reader)

	// Collect all events
	var collected []Event
	for event := range events {
		collected = append(collected, event)
	}

	require.Len(t, collected, 3)

	// First event: system init
	assert.Equal(t, EventTypeSystem, collected[0].Type)
	assert.True(t, collected[0].SessionStarted)

	// Second event: assistant text
	assert.Equal(t, EventTypeAssistant, collected[1].Type)
	assert.Equal(t, "Hello!", collected[1].Text)

	// Third event: result
	assert.Equal(t, EventTypeResult, collected[2].Type)
	assert.True(t, collected[2].SessionComplete)
}

func TestDefaultParser_Parse_SkipsInvalidJSON(t *testing.T) {
	input := `{"type":"system","subtype":"init"}
not valid json
{"type":"result"}`

	parser := NewParser()
	reader := strings.NewReader(input)

	events := parser.Parse(reader)

	// Collect all events
	var collected []Event
	for event := range events {
		collected = append(collected, event)
	}

	// Should have 2 events, skipping the invalid line
	require.Len(t, collected, 2)
	assert.Equal(t, EventTypeSystem, collected[0].Type)
	assert.Equal(t, EventTypeResult, collected[1].Type)
}

func TestDefaultParser_Parse_EmptyLines(t *testing.T) {
	input := `{"type":"system","subtype":"init"}

{"type":"result"}
`

	parser := NewParser()
	reader := strings.NewReader(input)

	events := parser.Parse(reader)

	// Collect all events
	var collected []Event
	for event := range events {
		collected = append(collected, event)
	}

	// Should have 2 events, skipping empty lines
	require.Len(t, collected, 2)
}

func TestDefaultParser_Parse_ToolUse(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la","description":"List files"}}]}}`

	parser := NewParser()
	reader := strings.NewReader(input)

	events := parser.Parse(reader)

	event := <-events
	assert.Equal(t, "Bash", event.ToolName)
	assert.Equal(t, "ls -la", event.ToolCommand)
	assert.Equal(t, "List files", event.ToolDescription)
}

func TestDefaultParser_Parse_ToolResult(t *testing.T) {
	input := `{"type":"user","tool_use_result":{"stdout":"file1.go\nfile2.go","stderr":""}}`

	parser := NewParser()
	reader := strings.NewReader(input)

	events := parser.Parse(reader)

	event := <-events
	assert.Equal(t, EventTypeUser, event.Type)
	assert.Equal(t, "file1.go\nfile2.go", event.ToolStdout)
	assert.True(t, event.IsToolResult())
}

func TestParseSingle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, event Event)
	}{
		{
			name:    "valid system event",
			input:   `{"type":"system","subtype":"init"}`,
			wantErr: false,
			check: func(t *testing.T, event Event) {
				assert.Equal(t, EventTypeSystem, event.Type)
				assert.True(t, event.SessionStarted)
			},
		},
		{
			name:    "invalid json",
			input:   `not json`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseSingle(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, event)
				}
			}
		})
	}
}
