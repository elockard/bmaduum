package lifecycle

import (
	"context"
	"errors"
	"testing"

	"bmaduum/internal/router"
	"bmaduum/internal/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWorkflowRunner implements WorkflowRunner for testing.
type MockWorkflowRunner struct {
	// RunSingleFunc allows tests to control workflow execution behavior.
	RunSingleFunc func(ctx context.Context, workflowName, storyKey string) int
	// Calls records all RunSingle calls for verification.
	Calls []struct {
		WorkflowName string
		StoryKey     string
	}
}

func (m *MockWorkflowRunner) RunSingle(ctx context.Context, workflowName, storyKey string) int {
	m.Calls = append(m.Calls, struct {
		WorkflowName string
		StoryKey     string
	}{workflowName, storyKey})

	if m.RunSingleFunc != nil {
		return m.RunSingleFunc(ctx, workflowName, storyKey)
	}
	return 0 // success by default
}

// MockStatusReader implements StatusReader for testing.
type MockStatusReader struct {
	// GetStoryStatusFunc allows tests to control status reading behavior.
	GetStoryStatusFunc func(storyKey string) (status.Status, error)
}

func (m *MockStatusReader) GetStoryStatus(storyKey string) (status.Status, error) {
	if m.GetStoryStatusFunc != nil {
		return m.GetStoryStatusFunc(storyKey)
	}
	return status.StatusBacklog, nil
}

// MockStatusWriter implements StatusWriter for testing.
type MockStatusWriter struct {
	// UpdateStatusFunc allows tests to control status writing behavior.
	UpdateStatusFunc func(storyKey string, newStatus status.Status) error
	// Calls records all UpdateStatus calls for verification.
	Calls []struct {
		StoryKey  string
		NewStatus status.Status
	}
}

func (m *MockStatusWriter) UpdateStatus(storyKey string, newStatus status.Status) error {
	m.Calls = append(m.Calls, struct {
		StoryKey  string
		NewStatus status.Status
	}{storyKey, newStatus})

	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(storyKey, newStatus)
	}
	return nil // success by default
}

func TestNewExecutor(t *testing.T) {
	runner := &MockWorkflowRunner{}
	reader := &MockStatusReader{}
	writer := &MockStatusWriter{}

	executor := NewExecutor(runner, reader, writer)

	assert.NotNil(t, executor)
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		storyKey       string
		currentStatus  status.Status
		getStatusErr   error
		workflowResult int
		updateErr      error
		wantErr        error
		wantWorkflows  []string
		wantUpdates    []status.Status
	}{
		{
			name:          "story in backlog runs full lifecycle",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusBacklog,
			wantWorkflows: []string{"create-story", "dev-story", "code-review", "git-commit"},
			wantUpdates:   []status.Status{status.StatusReadyForDev, status.StatusReview, status.StatusDone, status.StatusDone},
		},
		{
			name:          "story in ready-for-dev skips create-story",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusReadyForDev,
			wantWorkflows: []string{"dev-story", "code-review", "git-commit"},
			wantUpdates:   []status.Status{status.StatusReview, status.StatusDone, status.StatusDone},
		},
		{
			name:          "story in review runs code-review and git-commit",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusReview,
			wantWorkflows: []string{"code-review", "git-commit"},
			wantUpdates:   []status.Status{status.StatusDone, status.StatusDone},
		},
		{
			name:          "story already done returns ErrStoryComplete",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusDone,
			wantErr:       router.ErrStoryComplete,
			wantWorkflows: nil,
			wantUpdates:   nil,
		},
		{
			name:         "get status error propagates",
			storyKey:     "EPIC-1-story",
			getStatusErr: errors.New("file not found"),
			wantErr:      errors.New("file not found"),
		},
		{
			name:           "workflow failure stops execution",
			storyKey:       "EPIC-1-story",
			currentStatus:  status.StatusBacklog,
			workflowResult: 1, // first workflow fails
			wantWorkflows:  []string{"create-story"},
			wantUpdates:    nil, // no status updates if workflow fails
		},
		{
			name:          "status update failure stops execution",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusBacklog,
			updateErr:     errors.New("write failed"),
			wantWorkflows: []string{"create-story"},                  // workflow runs
			wantUpdates:   []status.Status{status.StatusReadyForDev}, // update attempted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockWorkflowRunner{
				RunSingleFunc: func(ctx context.Context, workflowName, storyKey string) int {
					return tt.workflowResult
				},
			}
			reader := &MockStatusReader{
				GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
					if tt.getStatusErr != nil {
						return "", tt.getStatusErr
					}
					return tt.currentStatus, nil
				},
			}
			writer := &MockStatusWriter{
				UpdateStatusFunc: func(storyKey string, newStatus status.Status) error {
					return tt.updateErr
				},
			}

			executor := NewExecutor(runner, reader, writer)
			err := executor.Execute(context.Background(), tt.storyKey)

			// Check error
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, router.ErrStoryComplete) {
					assert.ErrorIs(t, err, router.ErrStoryComplete)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
			} else if tt.workflowResult != 0 {
				// Workflow failure should return an error
				require.Error(t, err)
				assert.Contains(t, err.Error(), "workflow failed")
			} else if tt.updateErr != nil {
				// Update failure should return an error
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.updateErr.Error())
			} else {
				require.NoError(t, err)
			}

			// Check workflow calls
			if tt.wantWorkflows != nil {
				require.Len(t, runner.Calls, len(tt.wantWorkflows))
				for i, wantWorkflow := range tt.wantWorkflows {
					assert.Equal(t, wantWorkflow, runner.Calls[i].WorkflowName)
					assert.Equal(t, tt.storyKey, runner.Calls[i].StoryKey)
				}
			}

			// Check status updates
			if tt.wantUpdates != nil {
				require.Len(t, writer.Calls, len(tt.wantUpdates))
				for i, wantStatus := range tt.wantUpdates {
					assert.Equal(t, tt.storyKey, writer.Calls[i].StoryKey)
					assert.Equal(t, wantStatus, writer.Calls[i].NewStatus)
				}
			}
		})
	}
}

func TestProgressCallback(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus status.Status
		wantCalls     []struct {
			stepIndex  int
			totalSteps int
			workflow   string
		}
	}{
		{
			name:          "backlog story calls callback 4 times",
			currentStatus: status.StatusBacklog,
			wantCalls: []struct {
				stepIndex  int
				totalSteps int
				workflow   string
			}{
				{1, 4, "create-story"},
				{2, 4, "dev-story"},
				{3, 4, "code-review"},
				{4, 4, "git-commit"},
			},
		},
		{
			name:          "ready-for-dev story calls callback 3 times",
			currentStatus: status.StatusReadyForDev,
			wantCalls: []struct {
				stepIndex  int
				totalSteps int
				workflow   string
			}{
				{1, 3, "dev-story"},
				{2, 3, "code-review"},
				{3, 3, "git-commit"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track callback invocations
			var callbackCalls []struct {
				stepIndex  int
				totalSteps int
				workflow   string
			}

			runner := &MockWorkflowRunner{}
			reader := &MockStatusReader{
				GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
					return tt.currentStatus, nil
				},
			}
			writer := &MockStatusWriter{}

			executor := NewExecutor(runner, reader, writer)
			executor.SetProgressCallback(func(stepIndex, totalSteps int, workflow string) {
				callbackCalls = append(callbackCalls, struct {
					stepIndex  int
					totalSteps int
					workflow   string
				}{stepIndex, totalSteps, workflow})
			})

			err := executor.Execute(context.Background(), "test-story")
			require.NoError(t, err)

			// Verify callback was called with correct arguments
			require.Len(t, callbackCalls, len(tt.wantCalls))
			for i, want := range tt.wantCalls {
				assert.Equal(t, want.stepIndex, callbackCalls[i].stepIndex, "stepIndex mismatch at call %d", i)
				assert.Equal(t, want.totalSteps, callbackCalls[i].totalSteps, "totalSteps mismatch at call %d", i)
				assert.Equal(t, want.workflow, callbackCalls[i].workflow, "workflow mismatch at call %d", i)
			}
		})
	}
}

func TestProgressCallbackNotSet(t *testing.T) {
	// Verify that execution works without a callback set (no panic)
	runner := &MockWorkflowRunner{}
	reader := &MockStatusReader{
		GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
			return status.StatusBacklog, nil
		},
	}
	writer := &MockStatusWriter{}

	executor := NewExecutor(runner, reader, writer)
	// Do NOT set progress callback

	err := executor.Execute(context.Background(), "test-story")
	require.NoError(t, err)
	assert.Len(t, runner.Calls, 4) // All 4 workflows should run
}

// MockBmadHelpFallback implements BmadHelpFallback for testing.
type MockBmadHelpFallback struct {
	Workflow   string
	NextStatus status.Status
	Err        error
	Calls      []struct {
		StoryKey      string
		CurrentStatus status.Status
	}
}

func (m *MockBmadHelpFallback) ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (string, status.Status, error) {
	m.Calls = append(m.Calls, struct {
		StoryKey      string
		CurrentStatus status.Status
	}{storyKey, currentStatus})

	if m.Err != nil {
		return "", "", m.Err
	}
	return m.Workflow, m.NextStatus, nil
}

func TestExecute_BmadHelpFallback(t *testing.T) {
	t.Run("unknown status with bmad-help resolves and continues lifecycle", func(t *testing.T) {
		// Story starts with unknown status "pending-qa", bmad-help recommends code-review.
		// After code-review, status becomes "review" (recognized), so lifecycle continues normally.
		callCount := 0
		reader := &MockStatusReader{
			GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
				callCount++
				if callCount == 1 {
					return status.Status("pending-qa"), nil // unknown status
				}
				return status.StatusReview, nil // after bmad-help step, now at review
			},
		}
		runner := &MockWorkflowRunner{}
		writer := &MockStatusWriter{}
		bmadHelp := &MockBmadHelpFallback{
			Workflow:   "code-review",
			NextStatus: status.StatusReview,
		}

		executor := NewExecutor(runner, reader, writer)
		executor.SetBmadHelp(bmadHelp)

		err := executor.Execute(context.Background(), "STORY-1")
		require.NoError(t, err)

		// bmad-help should have been called once
		require.Len(t, bmadHelp.Calls, 1)
		assert.Equal(t, "STORY-1", bmadHelp.Calls[0].StoryKey)
		assert.Equal(t, status.Status("pending-qa"), bmadHelp.Calls[0].CurrentStatus)

		// Should have run: code-review (from bmad-help) + code-review + git-commit (from normal routing)
		require.Len(t, runner.Calls, 3)
		assert.Equal(t, "code-review", runner.Calls[0].WorkflowName) // bmad-help step
		assert.Equal(t, "code-review", runner.Calls[1].WorkflowName) // normal routing
		assert.Equal(t, "git-commit", runner.Calls[2].WorkflowName)  // normal routing

		// Status updates: review (bmad-help), done (code-review), done (git-commit)
		require.Len(t, writer.Calls, 3)
		assert.Equal(t, status.StatusReview, writer.Calls[0].NewStatus)
		assert.Equal(t, status.StatusDone, writer.Calls[1].NewStatus)
		assert.Equal(t, status.StatusDone, writer.Calls[2].NewStatus)
	})

	t.Run("unknown status with bmad-help resolves to done", func(t *testing.T) {
		// Story starts with unknown status, bmad-help recommends git-commit with done.
		// After git-commit, story is done.
		callCount := 0
		reader := &MockStatusReader{
			GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
				callCount++
				if callCount == 1 {
					return status.Status("awaiting-merge"), nil // unknown status
				}
				return status.StatusDone, nil // after bmad-help step, now done
			},
		}
		runner := &MockWorkflowRunner{}
		writer := &MockStatusWriter{}
		bmadHelp := &MockBmadHelpFallback{
			Workflow:   "git-commit",
			NextStatus: status.StatusDone,
		}

		executor := NewExecutor(runner, reader, writer)
		executor.SetBmadHelp(bmadHelp)

		err := executor.Execute(context.Background(), "STORY-1")
		// Should get ErrStoryComplete from the recursive call
		assert.ErrorIs(t, err, router.ErrStoryComplete)

		// bmad-help was called once
		require.Len(t, bmadHelp.Calls, 1)

		// git-commit was executed
		require.Len(t, runner.Calls, 1)
		assert.Equal(t, "git-commit", runner.Calls[0].WorkflowName)
	})

	t.Run("unknown status without bmad-help returns error", func(t *testing.T) {
		reader := &MockStatusReader{
			GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
				return status.Status("weird-status"), nil
			},
		}
		runner := &MockWorkflowRunner{}
		writer := &MockStatusWriter{}

		executor := NewExecutor(runner, reader, writer)
		// No bmad-help set

		err := executor.Execute(context.Background(), "STORY-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, router.ErrUnknownStatus)
		assert.Empty(t, runner.Calls)
	})

	t.Run("bmad-help failure returns wrapped error", func(t *testing.T) {
		reader := &MockStatusReader{
			GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
				return status.Status("weird-status"), nil
			},
		}
		runner := &MockWorkflowRunner{}
		writer := &MockStatusWriter{}
		bmadHelp := &MockBmadHelpFallback{
			Err: errors.New("Claude CLI not found"),
		}

		executor := NewExecutor(runner, reader, writer)
		executor.SetBmadHelp(bmadHelp)

		err := executor.Execute(context.Background(), "STORY-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bmad-help fallback failed")
		assert.Contains(t, err.Error(), "Claude CLI not found")
		assert.Empty(t, runner.Calls)
	})

	t.Run("depth limit prevents infinite recursion", func(t *testing.T) {
		// bmad-help always returns an unknown status, triggering recursion
		reader := &MockStatusReader{
			GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
				return status.Status("always-unknown"), nil
			},
		}
		runner := &MockWorkflowRunner{}
		writer := &MockStatusWriter{}
		bmadHelp := &MockBmadHelpFallback{
			Workflow:   "dev-story",
			NextStatus: status.Status("still-unknown"),
		}

		executor := NewExecutor(runner, reader, writer)
		executor.SetBmadHelp(bmadHelp)

		err := executor.Execute(context.Background(), "STORY-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum depth")

		// bmad-help should have been called maxBmadHelpDepth times
		assert.Len(t, bmadHelp.Calls, maxBmadHelpDepth)
	})
}

func TestGetSteps(t *testing.T) {
	tests := []struct {
		name          string
		storyKey      string
		currentStatus status.Status
		getStatusErr  error
		wantSteps     []router.LifecycleStep
		wantErr       error
	}{
		{
			name:          "story in backlog returns all 4 steps",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusBacklog,
			wantSteps: []router.LifecycleStep{
				{Workflow: "create-story", NextStatus: status.StatusReadyForDev},
				{Workflow: "dev-story", NextStatus: status.StatusReview},
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:          "story in ready-for-dev returns 3 steps",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusReadyForDev,
			wantSteps: []router.LifecycleStep{
				{Workflow: "dev-story", NextStatus: status.StatusReview},
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:          "story in review returns 2 steps",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusReview,
			wantSteps: []router.LifecycleStep{
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:          "story already done returns ErrStoryComplete",
			storyKey:      "EPIC-1-story",
			currentStatus: status.StatusDone,
			wantErr:       router.ErrStoryComplete,
		},
		{
			name:         "get status error propagates",
			storyKey:     "unknown-story",
			getStatusErr: errors.New("file not found"),
			wantErr:      errors.New("file not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetSteps should NOT call runner or writer
			runner := &MockWorkflowRunner{}
			reader := &MockStatusReader{
				GetStoryStatusFunc: func(storyKey string) (status.Status, error) {
					if tt.getStatusErr != nil {
						return "", tt.getStatusErr
					}
					return tt.currentStatus, nil
				},
			}
			writer := &MockStatusWriter{}

			executor := NewExecutor(runner, reader, writer)
			steps, err := executor.GetSteps(tt.storyKey)

			// Check error
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, router.ErrStoryComplete) {
					assert.ErrorIs(t, err, router.ErrStoryComplete)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				assert.Nil(t, steps)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSteps, steps)
			}

			// Verify no workflows were executed
			assert.Empty(t, runner.Calls, "GetSteps should not execute any workflows")

			// Verify no status updates were made
			assert.Empty(t, writer.Calls, "GetSteps should not update any status")
		})
	}
}
