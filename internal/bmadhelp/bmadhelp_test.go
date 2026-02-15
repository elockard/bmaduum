package bmadhelp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bmaduum/internal/claude"
	"bmaduum/internal/status"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		wantWorkflow string
		wantStatus   status.Status
		wantErr      bool
	}{
		{
			name:         "mentions dev-story directly",
			response:     "You should run dev-story for this story.",
			wantWorkflow: "dev-story",
			wantStatus:   status.StatusReview,
		},
		{
			name:         "mentions create-story",
			response:     "The next step is to run create-story to set up the story.",
			wantWorkflow: "create-story",
			wantStatus:   status.StatusReadyForDev,
		},
		{
			name:         "mentions code-review",
			response:     "I recommend running code-review next.",
			wantWorkflow: "code-review",
			wantStatus:   status.StatusDone,
		},
		{
			name:         "mentions git-commit",
			response:     "The story needs a git-commit to finalize changes.",
			wantWorkflow: "git-commit",
			wantStatus:   status.StatusDone,
		},
		{
			name:         "mentions test-automation",
			response:     "Run test-automation to verify the implementation.",
			wantWorkflow: "test-automation",
			wantStatus:   status.StatusDone,
		},
		{
			name:         "case insensitive match",
			response:     "You should run DEV-STORY for this.",
			wantWorkflow: "dev-story",
			wantStatus:   status.StatusReview,
		},
		{
			name:         "multiple workflows mentioned prefers first in chain",
			response:     "Run create-story first, then dev-story, then code-review.",
			wantWorkflow: "create-story",
			wantStatus:   status.StatusReadyForDev,
		},
		{
			name:         "workflow embedded in longer text",
			response:     "Based on the current state, I suggest executing the dev-story workflow to continue development on this story. This will move it forward in the pipeline.",
			wantWorkflow: "dev-story",
			wantStatus:   status.StatusReview,
		},
		{
			name:    "no recognizable workflow",
			response: "I'm not sure what to do with this story. Please check the sprint status.",
			wantErr: true,
		},
		{
			name:    "empty response",
			response: "",
			wantErr: true,
		},
		{
			name:    "unrelated workflow names",
			response: "Try running deploy or build next.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := ParseResponse(tt.response)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, rec)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rec)
			assert.Equal(t, tt.wantWorkflow, rec.Workflow)
			assert.Equal(t, tt.wantStatus, rec.NextStatus)
		})
	}
}

func TestClaudeFallback_ResolveWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		events        []claude.Event
		exitCode      int
		execErr       error
		wantWorkflow  string
		wantStatus    status.Status
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "successful resolution with dev-story",
			events: []claude.Event{
				{Type: claude.EventTypeAssistant, Text: "Based on the status, you should run dev-story next."},
			},
			exitCode:     0,
			wantWorkflow: "dev-story",
			wantStatus:   status.StatusReview,
		},
		{
			name: "successful resolution with create-story",
			events: []claude.Event{
				{Type: claude.EventTypeAssistant, Text: "The story needs to be created first. Run create-story."},
			},
			exitCode:     0,
			wantWorkflow: "create-story",
			wantStatus:   status.StatusReadyForDev,
		},
		{
			name: "multi-event response",
			events: []claude.Event{
				{Type: claude.EventTypeAssistant, Text: "Looking at the status..."},
				{Type: claude.EventTypeAssistant, Text: " I recommend running code-review."},
			},
			exitCode:     0,
			wantWorkflow: "code-review",
			wantStatus:   status.StatusDone,
		},
		{
			name: "non-zero exit code",
			events: []claude.Event{
				{Type: claude.EventTypeAssistant, Text: "Error occurred"},
			},
			exitCode:      1,
			wantErr:       true,
			wantErrSubstr: "exit code 1",
		},
		{
			name:          "executor error",
			execErr:       errors.New("connection failed"),
			wantErr:       true,
			wantErrSubstr: "execution failed",
		},
		{
			name: "no recognizable workflow in response",
			events: []claude.Event{
				{Type: claude.EventTypeAssistant, Text: "I don't know what to do with this status."},
			},
			exitCode:      0,
			wantErr:       true,
			wantErrSubstr: "did not contain a recognizable workflow",
		},
		{
			name:          "empty response (no events)",
			events:        nil,
			exitCode:      0,
			wantErr:       true,
			wantErrSubstr: "did not contain a recognizable workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &claude.MockExecutor{
				Events:   tt.events,
				ExitCode: tt.exitCode,
				Error:    tt.execErr,
			}

			fallback := NewClaudeFallback(mock)
			workflow, nextStatus, err := fallback.ResolveWorkflow(context.Background(), "STORY-1", status.Status("custom-status"))

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, workflow)
				if tt.wantErrSubstr != "" {
					assert.Contains(t, err.Error(), tt.wantErrSubstr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantWorkflow, workflow)
			assert.Equal(t, tt.wantStatus, nextStatus)

			// Verify the prompt was sent
			require.Len(t, mock.RecordedPrompts, 1)
			assert.Contains(t, mock.RecordedPrompts[0], "/bmad-help")
			assert.Contains(t, mock.RecordedPrompts[0], "STORY-1")
			assert.Contains(t, mock.RecordedPrompts[0], "custom-status")
		})
	}
}

func TestClaudeFallback_PromptFormat(t *testing.T) {
	mock := &claude.MockExecutor{
		Events: []claude.Event{
			{Type: claude.EventTypeAssistant, Text: "Run dev-story next."},
		},
		ExitCode: 0,
	}

	fallback := NewClaudeFallback(mock)
	_, _, _ = fallback.ResolveWorkflow(context.Background(), "7-3-implement-auth", status.Status("pending-review"))

	require.Len(t, mock.RecordedPrompts, 1)
	prompt := mock.RecordedPrompts[0]

	// Verify prompt structure
	assert.True(t, len(prompt) > 0, "prompt should not be empty")
	assert.Contains(t, prompt, "/bmad-help")
	assert.Contains(t, prompt, "7-3-implement-auth")
	assert.Contains(t, prompt, "pending-review")
}

func TestMockFallback(t *testing.T) {
	t.Run("returns configured recommendation", func(t *testing.T) {
		mock := &MockFallback{
			Rec: &Recommendation{Workflow: "dev-story", NextStatus: status.StatusReview},
		}

		workflow, nextStatus, err := mock.ResolveWorkflow(context.Background(), "STORY-1", status.Status("custom"))
		require.NoError(t, err)
		assert.Equal(t, "dev-story", workflow)
		assert.Equal(t, status.StatusReview, nextStatus)

		require.Len(t, mock.Calls, 1)
		assert.Equal(t, "STORY-1", mock.Calls[0].StoryKey)
		assert.Equal(t, status.Status("custom"), mock.Calls[0].CurrentStatus)
	})

	t.Run("returns configured error", func(t *testing.T) {
		mock := &MockFallback{
			Err: errors.New("bmad-help unavailable"),
		}

		workflow, _, err := mock.ResolveWorkflow(context.Background(), "STORY-1", status.Status("custom"))
		require.Error(t, err)
		assert.Empty(t, workflow)
		assert.Contains(t, err.Error(), "bmad-help unavailable")
	})

	t.Run("errors when nothing configured", func(t *testing.T) {
		mock := &MockFallback{}

		workflow, _, err := mock.ResolveWorkflow(context.Background(), "STORY-1", status.Status("custom"))
		require.Error(t, err)
		assert.Empty(t, workflow)
	})
}
