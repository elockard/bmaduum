package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bmad-automate/internal/claude"
	"bmad-automate/internal/config"
	"bmad-automate/internal/output"
	"bmad-automate/internal/status"
	"bmad-automate/internal/workflow"
)

func setupQueueTestApp(tmpDir string) (*App, *claude.MockExecutor, *bytes.Buffer) {
	cfg := config.DefaultConfig()
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)
	mockExecutor := &claude.MockExecutor{
		Events: []claude.Event{
			{Type: claude.EventTypeSystem, SessionStarted: true},
			{Type: claude.EventTypeResult, SessionComplete: true},
		},
		ExitCode: 0,
	}
	runner := workflow.NewRunner(mockExecutor, printer, cfg)
	queue := workflow.NewQueueRunner(runner)
	statusReader := status.NewReader(tmpDir)

	return &App{
		Config:       cfg,
		Executor:     mockExecutor,
		Printer:      printer,
		Runner:       runner,
		Queue:        queue,
		StatusReader: statusReader,
	}, mockExecutor, buf
}

func TestQueueCommand_StatusBasedRouting(t *testing.T) {
	tests := []struct {
		name             string
		storyKey         string
		statusYAML       string
		expectedWorkflow string
		expectError      bool
		expectExitCode   int
	}{
		{
			name:     "backlog status routes to create-story",
			storyKey: "STORY-1",
			statusYAML: `development_status:
  STORY-1: backlog`,
			expectedWorkflow: "create-story",
			expectError:      false,
		},
		{
			name:     "ready-for-dev status routes to dev-story",
			storyKey: "STORY-2",
			statusYAML: `development_status:
  STORY-2: ready-for-dev`,
			expectedWorkflow: "dev-story",
			expectError:      false,
		},
		{
			name:     "in-progress status routes to dev-story",
			storyKey: "STORY-3",
			statusYAML: `development_status:
  STORY-3: in-progress`,
			expectedWorkflow: "dev-story",
			expectError:      false,
		},
		{
			name:     "review status routes to code-review",
			storyKey: "STORY-4",
			statusYAML: `development_status:
  STORY-4: review`,
			expectedWorkflow: "code-review",
			expectError:      false,
		},
		{
			name:     "done status is skipped",
			storyKey: "STORY-5",
			statusYAML: `development_status:
  STORY-5: done`,
			expectedWorkflow: "", // No workflow executed
			expectError:      false,
		},
		{
			name:     "story not found returns error",
			storyKey: "STORY-NOT-FOUND",
			statusYAML: `development_status:
  STORY-1: backlog`,
			expectError:    true,
			expectExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			createSprintStatusFile(t, tmpDir, tt.statusYAML)

			app, mockExecutor, _ := setupQueueTestApp(tmpDir)
			rootCmd := NewRootCommand(app)

			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(errBuf)
			rootCmd.SetArgs([]string{"queue", tt.storyKey})

			err := rootCmd.Execute()

			if tt.expectError {
				require.Error(t, err)
				if tt.expectExitCode > 0 {
					code, ok := IsExitError(err)
					assert.True(t, ok, "error should be an ExitError")
					assert.Equal(t, tt.expectExitCode, code)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedWorkflow != "" {
				assert.NotEmpty(t, mockExecutor.RecordedPrompts, "prompt should have been executed")
			}
		})
	}
}

func TestQueueCommand_DoneStorySkipped(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: done`)

	app, mockExecutor, _ := setupQueueTestApp(tmpDir)
	rootCmd := NewRootCommand(app)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"queue", "STORY-1"})

	err := rootCmd.Execute()

	assert.NoError(t, err, "done story should not cause error")
	assert.Empty(t, mockExecutor.RecordedPrompts, "no workflow should be executed for done story")
}

func TestQueueCommand_StopsOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog
  STORY-2: backlog`)

	cfg := config.DefaultConfig()
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)
	mockExecutor := &claude.MockExecutor{
		Events: []claude.Event{
			{Type: claude.EventTypeSystem, SessionStarted: true},
			{Type: claude.EventTypeResult, SessionComplete: true},
		},
		ExitCode: 1, // Simulate failure
	}
	runner := workflow.NewRunner(mockExecutor, printer, cfg)
	queue := workflow.NewQueueRunner(runner)
	statusReader := status.NewReader(tmpDir)

	app := &App{
		Config:       cfg,
		Executor:     mockExecutor,
		Printer:      printer,
		Runner:       runner,
		Queue:        queue,
		StatusReader: statusReader,
	}

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"queue", "STORY-1", "STORY-2"})

	err := rootCmd.Execute()

	require.Error(t, err)
	code, ok := IsExitError(err)
	assert.True(t, ok, "error should be an ExitError")
	assert.Equal(t, 1, code)

	// Should have only run one workflow (stopped on first failure)
	assert.Len(t, mockExecutor.RecordedPrompts, 1, "should stop after first failure")
}

func TestQueueCommand_MixedStatusQueue(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog
  STORY-2: done
  STORY-3: ready-for-dev`)

	app, mockExecutor, _ := setupQueueTestApp(tmpDir)
	rootCmd := NewRootCommand(app)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"queue", "STORY-1", "STORY-2", "STORY-3"})

	err := rootCmd.Execute()

	assert.NoError(t, err)
	// Should have run 2 workflows (STORY-1 and STORY-3), skipping STORY-2
	assert.Len(t, mockExecutor.RecordedPrompts, 2, "should run workflow for backlog and ready-for-dev, skip done")
}

func TestQueueCommand_MissingSprintStatusFile(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create sprint-status.yaml

	app, _, _ := setupQueueTestApp(tmpDir)
	rootCmd := NewRootCommand(app)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"queue", "STORY-1"})

	err := rootCmd.Execute()

	require.Error(t, err)
	code, ok := IsExitError(err)
	assert.True(t, ok, "error should be an ExitError")
	assert.Equal(t, 1, code)
}

func TestQueueCommand_MultipleStoriesAllSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog
  STORY-2: ready-for-dev
  STORY-3: review`)

	app, mockExecutor, _ := setupQueueTestApp(tmpDir)
	rootCmd := NewRootCommand(app)

	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"queue", "STORY-1", "STORY-2", "STORY-3"})

	err := rootCmd.Execute()

	assert.NoError(t, err)
	assert.Len(t, mockExecutor.RecordedPrompts, 3, "should run workflow for all three stories")
}
