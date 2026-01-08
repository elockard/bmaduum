// Package claude provides types and functionality for interacting with the Claude CLI.
package claude

// StreamEvent represents a JSON event from Claude's streaming output.
type StreamEvent struct {
	Type          string          `json:"type"`
	Subtype       string          `json:"subtype,omitempty"`
	Message       *MessageContent `json:"message,omitempty"`
	ToolUseResult *ToolResult     `json:"tool_use_result,omitempty"`
}

// MessageContent represents the content of a message from Claude.
type MessageContent struct {
	Content []ContentBlock `json:"content,omitempty"`
}

// ContentBlock represents a single block of content (text or tool use).
type ContentBlock struct {
	Type  string     `json:"type"`
	Text  string     `json:"text,omitempty"`
	Name  string     `json:"name,omitempty"`
	Input *ToolInput `json:"input,omitempty"`
}

// ToolInput represents the input parameters for a tool invocation.
type ToolInput struct {
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Content     string `json:"content,omitempty"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Stdout      string `json:"stdout,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	Interrupted bool   `json:"interrupted,omitempty"`
}

// EventType represents the type of event received from Claude.
type EventType string

const (
	EventTypeSystem    EventType = "system"
	EventTypeAssistant EventType = "assistant"
	EventTypeUser      EventType = "user"
	EventTypeResult    EventType = "result"
)

// SubtypeInit is the subtype for system initialization events.
const SubtypeInit = "init"

// Event is a parsed event from Claude's streaming output.
// It wraps StreamEvent and provides convenience methods.
type Event struct {
	Raw *StreamEvent

	// Parsed fields for convenience
	Type    EventType
	Subtype string

	// Text content (if Type == EventTypeAssistant and block is text)
	Text string

	// Tool use (if Type == EventTypeAssistant and block is tool_use)
	ToolName        string
	ToolDescription string
	ToolCommand     string
	ToolFilePath    string

	// Tool result (if Type == EventTypeUser and has tool_use_result)
	ToolStdout      string
	ToolStderr      string
	ToolInterrupted bool

	// Session state
	SessionStarted  bool
	SessionComplete bool
}

// NewEventFromStream creates an Event from a StreamEvent.
func NewEventFromStream(raw *StreamEvent) Event {
	e := Event{
		Raw:     raw,
		Type:    EventType(raw.Type),
		Subtype: raw.Subtype,
	}

	switch e.Type {
	case EventTypeSystem:
		if raw.Subtype == SubtypeInit {
			e.SessionStarted = true
		}

	case EventTypeAssistant:
		if raw.Message != nil {
			for _, block := range raw.Message.Content {
				switch block.Type {
				case "text":
					e.Text = block.Text
				case "tool_use":
					e.ToolName = block.Name
					if block.Input != nil {
						e.ToolDescription = block.Input.Description
						e.ToolCommand = block.Input.Command
						e.ToolFilePath = block.Input.FilePath
					}
				}
			}
		}

	case EventTypeUser:
		if raw.ToolUseResult != nil {
			e.ToolStdout = raw.ToolUseResult.Stdout
			e.ToolStderr = raw.ToolUseResult.Stderr
			e.ToolInterrupted = raw.ToolUseResult.Interrupted
		}

	case EventTypeResult:
		e.SessionComplete = true
	}

	return e
}

// IsText returns true if this event contains text content.
func (e Event) IsText() bool {
	return e.Type == EventTypeAssistant && e.Text != ""
}

// IsToolUse returns true if this event is a tool invocation.
func (e Event) IsToolUse() bool {
	return e.Type == EventTypeAssistant && e.ToolName != ""
}

// IsToolResult returns true if this event contains a tool result.
func (e Event) IsToolResult() bool {
	return e.Type == EventTypeUser && (e.ToolStdout != "" || e.ToolStderr != "")
}
