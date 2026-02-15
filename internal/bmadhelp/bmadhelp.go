// Package bmadhelp provides a fallback routing mechanism using BMAD's /bmad-help
// slash command when the standard router cannot determine the next workflow step.
//
// This is a last-resort fallback for edge cases such as unknown status values,
// missing stories, or ambiguous state. The primary routing path remains the
// standard router in internal/router.
//
// Key types:
//   - [Fallback] - Interface for resolving unknown statuses to workflow recommendations
//   - [ClaudeFallback] - Production implementation using Claude CLI via /bmad-help
//   - [MockFallback] - Test implementation with configurable responses
//   - [Recommendation] - The resolved workflow name and expected next status
package bmadhelp

import (
	"context"
	"fmt"
	"strings"

	"bmaduum/internal/claude"
	"bmaduum/internal/status"
)

// knownWorkflows is the set of standard workflow names that can be extracted
// from a /bmad-help response. Order matters: earlier entries are preferred
// when multiple workflow names appear in the response.
var knownWorkflows = []string{
	"create-story",
	"dev-story",
	"code-review",
	"test-automation",
	"git-commit",
}

// workflowNextStatus maps workflow names to their expected next status.
// Used when bmad-help doesn't explicitly provide a next status.
var workflowNextStatus = map[string]status.Status{
	"create-story":    status.StatusReadyForDev,
	"dev-story":       status.StatusReview,
	"code-review":     status.StatusDone,
	"test-automation": status.StatusDone,
	"git-commit":      status.StatusDone,
}

// Recommendation is the result of a bmad-help fallback resolution.
type Recommendation struct {
	// Workflow is the recommended workflow name to execute.
	Workflow string

	// NextStatus is the expected status after the workflow completes.
	NextStatus status.Status
}

// Fallback resolves unknown statuses to workflow recommendations via /bmad-help.
//
// Implementations invoke Claude CLI with the /bmad-help slash command and parse
// the response to extract a workflow recommendation. This is used as a last-resort
// fallback when the standard router returns [router.ErrUnknownStatus].
//
// Fallback also satisfies the lifecycle.BmadHelpFallback interface.
type Fallback interface {
	// ResolveWorkflow asks /bmad-help for the next workflow step for a story
	// with an unknown status. Returns the workflow name and expected next status.
	//
	// Returns an error if Claude CLI fails or the response cannot be parsed
	// into a recognizable workflow recommendation.
	ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (workflow string, nextStatus status.Status, err error)
}

// ClaudeFallback implements [Fallback] by invoking /bmad-help via Claude CLI.
//
// Create instances using [NewClaudeFallback]. The executor should be the same
// Claude executor used for workflow execution.
type ClaudeFallback struct {
	executor claude.Executor
}

// NewClaudeFallback creates a new [ClaudeFallback] with the given Claude executor.
func NewClaudeFallback(executor claude.Executor) *ClaudeFallback {
	return &ClaudeFallback{executor: executor}
}

// ResolveWorkflow invokes /bmad-help to determine the next workflow for a story
// with an unrecognized status value.
//
// The prompt asks bmad-help for routing guidance, and the response is parsed
// to find a known workflow name. If no recognizable workflow is found in the
// response, an error is returned.
func (f *ClaudeFallback) ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (string, status.Status, error) {
	prompt := fmt.Sprintf(
		`/bmad-help The story %s has status "%s" which is not a standard status. What is the next workflow step to run? Please respond with the workflow name (create-story, dev-story, code-review, test-automation, or git-commit).`,
		storyKey, currentStatus,
	)

	// Collect text from Claude's response
	var responseText strings.Builder
	handler := func(event claude.Event) {
		if event.IsText() {
			responseText.WriteString(event.Text)
		}
	}

	exitCode, err := f.executor.ExecuteWithResult(ctx, prompt, handler, "")
	if err != nil {
		return "", "", fmt.Errorf("bmad-help execution failed: %w", err)
	}
	if exitCode != 0 {
		return "", "", fmt.Errorf("bmad-help returned exit code %d", exitCode)
	}

	rec, err := ParseResponse(responseText.String())
	if err != nil {
		return "", "", err
	}
	return rec.Workflow, rec.NextStatus, nil
}

// ParseResponse extracts a workflow recommendation from a /bmad-help response.
//
// It scans the response text for known workflow names and returns the first
// match with its expected next status. Returns an error if no recognizable
// workflow name is found.
func ParseResponse(response string) (*Recommendation, error) {
	lower := strings.ToLower(response)

	for _, workflow := range knownWorkflows {
		if strings.Contains(lower, workflow) {
			nextStatus, ok := workflowNextStatus[workflow]
			if !ok {
				nextStatus = status.StatusDone
			}
			return &Recommendation{
				Workflow:   workflow,
				NextStatus: nextStatus,
			}, nil
		}
	}

	return nil, fmt.Errorf("bmad-help response did not contain a recognizable workflow recommendation")
}

// MockFallback implements [Fallback] for testing.
//
// Configure the mock by setting its fields before calling ResolveWorkflow:
//
//	mock := &MockFallback{
//	    Rec: &Recommendation{Workflow: "dev-story", NextStatus: status.StatusReview},
//	}
type MockFallback struct {
	// Rec is the recommendation to return. If nil and Err is nil, returns an error.
	Rec *Recommendation

	// Err is the error to return from ResolveWorkflow.
	Err error

	// Calls records all ResolveWorkflow invocations for verification.
	Calls []struct {
		StoryKey      string
		CurrentStatus status.Status
	}
}

// ResolveWorkflow returns the pre-configured recommendation or error.
func (m *MockFallback) ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (string, status.Status, error) {
	m.Calls = append(m.Calls, struct {
		StoryKey      string
		CurrentStatus status.Status
	}{storyKey, currentStatus})

	if m.Err != nil {
		return "", "", m.Err
	}
	if m.Rec != nil {
		return m.Rec.Workflow, m.Rec.NextStatus, nil
	}
	return "", "", fmt.Errorf("MockFallback: no recommendation or error configured")
}
