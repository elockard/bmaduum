package claude

import (
	"bufio"
	"encoding/json"
	"io"
)

// Parser parses streaming JSON output from Claude.
type Parser interface {
	// Parse reads from the given reader and returns a channel of events.
	// The channel is closed when the reader is exhausted or an error occurs.
	Parse(reader io.Reader) <-chan Event
}

// DefaultParser implements Parser for Claude's stream-json format.
type DefaultParser struct {
	// BufferSize is the maximum size for a single JSON line.
	// Defaults to 10MB if not set.
	BufferSize int
}

// NewParser creates a new DefaultParser with default settings.
func NewParser() *DefaultParser {
	return &DefaultParser{
		BufferSize: 10 * 1024 * 1024, // 10MB
	}
}

// Parse reads streaming JSON from the reader and emits parsed events.
func (p *DefaultParser) Parse(reader io.Reader) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		scanner := bufio.NewScanner(reader)

		// Set up buffer for large JSON lines
		bufSize := p.BufferSize
		if bufSize <= 0 {
			bufSize = 10 * 1024 * 1024
		}
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, bufSize)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var streamEvent StreamEvent
			if err := json.Unmarshal([]byte(line), &streamEvent); err != nil {
				// Skip unparseable lines
				continue
			}

			event := NewEventFromStream(&streamEvent)
			events <- event
		}

		// Note: scanner.Err() is intentionally not checked here
		// as we want to gracefully handle EOF and pipe closure
	}()

	return events
}

// ParseSingle parses a single JSON line into an Event.
// Returns an error if parsing fails.
func ParseSingle(line string) (Event, error) {
	var streamEvent StreamEvent
	if err := json.Unmarshal([]byte(line), &streamEvent); err != nil {
		return Event{}, err
	}
	return NewEventFromStream(&streamEvent), nil
}
