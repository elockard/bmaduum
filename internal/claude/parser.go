package claude

import (
	"bufio"
	"encoding/json"
	"io"
)

// Parser parses streaming JSON output from Claude CLI.
//
// The parser expects Claude's stream-json format, where each line of output is a
// complete JSON object representing a [StreamEvent]. The parser reads lines,
// deserializes them, and converts them to [Event] objects.
//
// The channel returned by Parse is closed when:
//   - EOF is reached (normal completion)
//   - The underlying reader is closed
//   - An unrecoverable read error occurs
//
// Malformed JSON lines are silently skipped to provide resilience against
// partial or corrupted output.
type Parser interface {
	// Parse reads streaming JSON from the given reader and returns a channel of [Event] objects.
	// The channel is closed when the reader is exhausted or an error occurs.
	// Empty lines and unparseable JSON lines are skipped.
	Parse(reader io.Reader) <-chan Event
}

// DefaultParser implements [Parser] for Claude's stream-json format.
//
// DefaultParser uses a buffered scanner to read JSON lines from Claude's stdout.
// The BufferSize field controls the maximum allowed line length, which is important
// because Claude may output large JSON objects (e.g., file contents in tool results).
//
// Create instances using [NewParser] rather than constructing directly to ensure
// proper default values.
type DefaultParser struct {
	// BufferSize is the maximum size in bytes for a single JSON line.
	// Lines exceeding this size will cause a scanner error and stop parsing.
	// Defaults to 10MB (10 * 1024 * 1024) if not set or <= 0.
	BufferSize int
}

// NewParser creates a new [DefaultParser] with default settings.
//
// The default buffer size is 10MB, which should accommodate most Claude output
// including large file contents in tool results.
func NewParser() *DefaultParser {
	return &DefaultParser{
		BufferSize: 10 * 1024 * 1024, // 10MB
	}
}

// Parse reads streaming JSON from the reader and emits parsed [Event] objects.
//
// Parse spawns a goroutine that reads lines from the reader, parses each line as
// a [StreamEvent], converts it to an [Event], and sends it to the returned channel.
//
// Error handling behavior:
//   - Empty lines are silently skipped
//   - Lines that fail JSON parsing are silently skipped (resilience against partial output)
//   - Scanner errors (e.g., line too long) terminate parsing and close the channel
//   - EOF closes the channel normally
//
// The scanner buffer is configured based on [DefaultParser.BufferSize] to handle
// large JSON objects that may appear in Claude's output.
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

// ParseSingle parses a single JSON line into an [Event].
//
// This is a utility function useful for testing and debugging. It parses a single
// line of Claude's stream-json output without requiring a reader or channel.
//
// Returns an error if the JSON is malformed or cannot be unmarshaled into a
// [StreamEvent]. Unlike [Parser.Parse], this function does not silently skip
// invalid input.
//
// Example:
//
//	event, err := ParseSingle(`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(event.Text) // "Hello"
func ParseSingle(line string) (Event, error) {
	var streamEvent StreamEvent
	if err := json.Unmarshal([]byte(line), &streamEvent); err != nil {
		return Event{}, err
	}
	return NewEventFromStream(&streamEvent), nil
}
