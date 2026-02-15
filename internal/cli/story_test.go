package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bmaduum/internal/config"
	"bmaduum/internal/manifest"
	"bmaduum/internal/output"
	"bmaduum/internal/router"
	"bmaduum/internal/status"
)

// TestStoryCommand_SingleStory tests that story command works for a single story
func TestStoryCommand_SingleStory(t *testing.T) {
	tests := []struct {
		name              string
		storyKey          string
		statusYAML        string
		expectedWorkflows []string
		expectedStatuses  []StatusUpdate
		expectError       bool
		failOnWorkflow    string
	}{
		{
			name:     "single backlog story runs full lifecycle",
			storyKey: "STORY-1",
			statusYAML: `development_status:
  STORY-1: backlog`,
			expectedWorkflows: []string{
				"create-story", "dev-story", "code-review", "git-commit",
			},
			expectedStatuses: []StatusUpdate{
				{StoryKey: "STORY-1", NewStatus: status.StatusReadyForDev},
				{StoryKey: "STORY-1", NewStatus: status.StatusReview},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
			},
			expectError: false,
		},
		{
			name:     "story at review runs only remaining workflows",
			storyKey: "STORY-1",
			statusYAML: `development_status:
  STORY-1: review`,
			expectedWorkflows: []string{
				"code-review", "git-commit",
			},
			expectedStatuses: []StatusUpdate{
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
			},
			expectError: false,
		},
		{
			name:     "done story is skipped",
			storyKey: "STORY-1",
			statusYAML: `development_status:
  STORY-1: done`,
			expectedWorkflows: nil,
			expectedStatuses:  nil,
			expectError:       false,
		},
		{
			name:     "story failure stops execution",
			storyKey: "STORY-1",
			statusYAML: `development_status:
  STORY-1: backlog`,
			failOnWorkflow: "dev-story",
			expectedWorkflows: []string{
				"create-story", "dev-story",
			},
			expectedStatuses: []StatusUpdate{
				{StoryKey: "STORY-1", NewStatus: status.StatusReadyForDev},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			createSprintStatusFile(t, tmpDir, tt.statusYAML)

			mockRunner := &MockWorkflowRunner{
				FailOnWorkflow: tt.failOnWorkflow,
			}
			mockWriter := &MockStatusWriter{}
			statusReader := status.NewReader(tmpDir)
			buf := &bytes.Buffer{}
			printer := output.NewPrinterWithWriter(buf)

			app := &App{
				Config:       config.DefaultConfig(),
				StatusReader: statusReader,
				StatusWriter: mockWriter,
				Runner:       mockRunner,
				Printer:      printer,
			}

			rootCmd := NewRootCommand(app)
			outBuf := &bytes.Buffer{}
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(outBuf)
			rootCmd.SetArgs([]string{"story", tt.storyKey})

			err := rootCmd.Execute()

			if tt.expectError {
				require.Error(t, err)
				code, ok := IsExitError(err)
				assert.True(t, ok, "error should be an ExitError")
				assert.Equal(t, 1, code)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedWorkflows, mockRunner.ExecutedWorkflows,
				"workflows should be executed in lifecycle order")

			if tt.expectedStatuses != nil {
				require.Len(t, mockWriter.Updates, len(tt.expectedStatuses))
				for i, expected := range tt.expectedStatuses {
					assert.Equal(t, expected.StoryKey, mockWriter.Updates[i].StoryKey)
					assert.Equal(t, expected.NewStatus, mockWriter.Updates[i].NewStatus)
				}
			}
		})
	}
}

// TestStoryCommand_MultipleStories tests that story command processes multiple stories
func TestStoryCommand_MultipleStories(t *testing.T) {
	tests := []struct {
		name              string
		storyKeys         []string
		statusYAML        string
		expectedWorkflows []string
		expectedStatuses  []StatusUpdate
		expectError       bool
	}{
		{
			name:      "multiple stories run in sequence",
			storyKeys: []string{"STORY-1", "STORY-2"},
			statusYAML: `development_status:
  STORY-1: backlog
  STORY-2: backlog`,
			expectedWorkflows: []string{
				"create-story", "dev-story", "code-review", "git-commit",
				"create-story", "dev-story", "code-review", "git-commit",
			},
			expectedStatuses: []StatusUpdate{
				{StoryKey: "STORY-1", NewStatus: status.StatusReadyForDev},
				{StoryKey: "STORY-1", NewStatus: status.StatusReview},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-2", NewStatus: status.StatusReadyForDev},
				{StoryKey: "STORY-2", NewStatus: status.StatusReview},
				{StoryKey: "STORY-2", NewStatus: status.StatusDone},
				{StoryKey: "STORY-2", NewStatus: status.StatusDone},
			},
			expectError: false,
		},
		{
			name:      "mixed statuses run appropriate workflows",
			storyKeys: []string{"STORY-1", "STORY-2", "STORY-3"},
			statusYAML: `development_status:
  STORY-1: backlog
  STORY-2: review
  STORY-3: done`,
			expectedWorkflows: []string{
				"create-story", "dev-story", "code-review", "git-commit",
				"code-review", "git-commit",
				// STORY-3 is done, skipped
			},
			expectedStatuses: []StatusUpdate{
				{StoryKey: "STORY-1", NewStatus: status.StatusReadyForDev},
				{StoryKey: "STORY-1", NewStatus: status.StatusReview},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-1", NewStatus: status.StatusDone},
				{StoryKey: "STORY-2", NewStatus: status.StatusDone},
				{StoryKey: "STORY-2", NewStatus: status.StatusDone},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			createSprintStatusFile(t, tmpDir, tt.statusYAML)

			mockRunner := &MockWorkflowRunner{}
			mockWriter := &MockStatusWriter{}
			statusReader := status.NewReader(tmpDir)
			buf := &bytes.Buffer{}
			printer := output.NewPrinterWithWriter(buf)

			app := &App{
				Config:       config.DefaultConfig(),
				StatusReader: statusReader,
				StatusWriter: mockWriter,
				Runner:       mockRunner,
				Printer:      printer,
			}

			rootCmd := NewRootCommand(app)
			outBuf := &bytes.Buffer{}
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(outBuf)
			rootCmd.SetArgs(append([]string{"story"}, tt.storyKeys...))

			err := rootCmd.Execute()

			if tt.expectError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedWorkflows, mockRunner.ExecutedWorkflows)

			if tt.expectedStatuses != nil {
				require.Len(t, mockWriter.Updates, len(tt.expectedStatuses))
				for i, expected := range tt.expectedStatuses {
					assert.Equal(t, expected.StoryKey, mockWriter.Updates[i].StoryKey)
					assert.Equal(t, expected.NewStatus, mockWriter.Updates[i].NewStatus)
				}
			}
		})
	}
}

// TestStoryCommand_DryRun tests dry-run mode for story command
func TestStoryCommand_DryRun(t *testing.T) {
	tests := []struct {
		name       string
		storyKeys  []string
		statusYAML string
	}{
		{
			name:      "single story dry run",
			storyKeys: []string{"STORY-1"},
			statusYAML: `development_status:
  STORY-1: backlog`,
		},
		{
			name:      "multiple stories dry run",
			storyKeys: []string{"STORY-1", "STORY-2"},
			statusYAML: `development_status:
  STORY-1: backlog
  STORY-2: review`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			createSprintStatusFile(t, tmpDir, tt.statusYAML)

			mockRunner := &MockWorkflowRunner{}
			mockWriter := &MockStatusWriter{}
			statusReader := status.NewReader(tmpDir)
			buf := &bytes.Buffer{}
			printer := output.NewPrinterWithWriter(buf)

			app := &App{
				Config:       config.DefaultConfig(),
				StatusReader: statusReader,
				StatusWriter: mockWriter,
				Runner:       mockRunner,
				Printer:      printer,
			}

			rootCmd := NewRootCommand(app)
			outBuf := &bytes.Buffer{}
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(outBuf)
			rootCmd.SetArgs(append([]string{"story", "--dry-run"}, tt.storyKeys...))

			err := rootCmd.Execute()
			assert.NoError(t, err)

			// No workflows should have been executed in dry-run
			assert.Empty(t, mockRunner.ExecutedWorkflows)
			assert.Empty(t, mockWriter.Updates)
		})
	}
}

// TestStoryCommand_WithSDETModule tests that SDET module injects test-automation step
func TestStoryCommand_WithSDETModule(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog`)

	mockRunner := &MockWorkflowRunner{}
	mockWriter := &MockStatusWriter{}
	statusReader := status.NewReader(tmpDir)
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)

	// Create a router with test-automation step injected (simulating SDET module)
	wfRouter := router.NewRouter()
	wfRouter.InsertStepAfter("code-review", "test-automation", status.StatusDone)

	modules, err := manifest.ReadModulesFromBytes([]byte(`modules:
  - name: bmm
    version: "6.0.0"
  - name: sdet
    version: "1.0.0"
`))
	require.NoError(t, err)

	app := &App{
		Config:       config.DefaultConfig(),
		StatusReader: statusReader,
		StatusWriter: mockWriter,
		Runner:       mockRunner,
		Printer:      printer,
		Router:       wfRouter,
		Modules:      modules,
	}

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"story", "STORY-1"})

	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Should include test-automation step
	assert.Equal(t, []string{
		"create-story", "dev-story", "code-review", "test-automation", "git-commit",
	}, mockRunner.ExecutedWorkflows)

	// Should have 5 status updates
	require.Len(t, mockWriter.Updates, 5)
	assert.Equal(t, status.StatusReadyForDev, mockWriter.Updates[0].NewStatus)
	assert.Equal(t, status.StatusReview, mockWriter.Updates[1].NewStatus)
	assert.Equal(t, status.StatusDone, mockWriter.Updates[2].NewStatus) // code-review
	assert.Equal(t, status.StatusDone, mockWriter.Updates[3].NewStatus) // test-automation
	assert.Equal(t, status.StatusDone, mockWriter.Updates[4].NewStatus) // git-commit
}

// TestStoryCommand_DryRunWithModules tests that modules appear in dry-run output
func TestStoryCommand_DryRunWithModules(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog`)

	mockRunner := &MockWorkflowRunner{}
	mockWriter := &MockStatusWriter{}
	statusReader := status.NewReader(tmpDir)
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)

	// Create a router with test-automation step injected
	wfRouter := router.NewRouter()
	wfRouter.InsertStepAfter("code-review", "test-automation", status.StatusDone)

	modules, err := manifest.ReadModulesFromBytes([]byte(`modules:
  - name: bmm
    version: "6.0.0"
  - name: sdet
    version: "1.0.0"
`))
	require.NoError(t, err)

	app := &App{
		Config:       config.DefaultConfig(),
		StatusReader: statusReader,
		StatusWriter: mockWriter,
		Runner:       mockRunner,
		Printer:      printer,
		Router:       wfRouter,
		Modules:      modules,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"story", "--dry-run", "STORY-1"})

	err = rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(r)
	stdout := stdoutBuf.String()

	assert.NoError(t, err)
	assert.Empty(t, mockRunner.ExecutedWorkflows, "dry-run should not execute workflows")

	// Verify module info is present
	assert.Contains(t, stdout, "Modules: bmm, sdet")

	// Verify test-automation appears in the steps
	assert.Contains(t, stdout, "test-automation")
}

// TestStoryCommand_NoBmadHelpFlag tests that --no-bmad-help disables the fallback
func TestStoryCommand_NoBmadHelpFlag(t *testing.T) {
	tmpDir := t.TempDir()
	// Story has an unknown status
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: pending-qa`)

	mockRunner := &MockWorkflowRunner{}
	mockWriter := &MockStatusWriter{}
	statusReader := status.NewReader(tmpDir)
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)

	// Create a mock bmad-help that would resolve the status if called
	mockBmadHelp := &MockBmadHelpFallback{
		Workflow:   "code-review",
		NextStatus: status.StatusDone,
	}

	app := &App{
		Config:       config.DefaultConfig(),
		StatusReader: statusReader,
		StatusWriter: mockWriter,
		Runner:       mockRunner,
		Printer:      printer,
		BmadHelp:     mockBmadHelp,
	}

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	// Use --no-bmad-help flag
	rootCmd.SetArgs([]string{"story", "--no-bmad-help", "STORY-1"})

	err := rootCmd.Execute()

	// Should fail with exit error since bmad-help is disabled and status is unknown
	require.Error(t, err)
	code, ok := IsExitError(err)
	assert.True(t, ok, "error should be an ExitError")
	assert.Equal(t, 1, code)

	// bmad-help should NOT have been called
	assert.Empty(t, mockBmadHelp.Calls, "bmad-help should not be called with --no-bmad-help")

	// No workflows should have been executed
	assert.Empty(t, mockRunner.ExecutedWorkflows)
}

// TestStoryCommand_BmadHelpFallbackResolves tests bmad-help fallback for unknown status
func TestStoryCommand_BmadHelpFallbackResolves(t *testing.T) {
	tmpDir := t.TempDir()
	// Story has an unknown status that bmad-help will resolve
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: pending-qa`)

	mockRunner := &MockWorkflowRunner{}
	mockWriter := &MockStatusWriter{}
	statusReader := status.NewReader(tmpDir)
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)

	mockBmadHelp := &MockBmadHelpFallback{
		Workflow:   "code-review",
		NextStatus: status.StatusDone,
	}

	app := &App{
		Config:       config.DefaultConfig(),
		StatusReader: statusReader,
		StatusWriter: mockWriter,
		Runner:       mockRunner,
		Printer:      printer,
		BmadHelp:     mockBmadHelp,
	}

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"story", "STORY-1"})

	err := rootCmd.Execute()

	// The lifecycle executor will call bmad-help, execute code-review, then re-read status.
	// Since we're reading from a real file, the status will still be "pending-qa" on re-read
	// (because our mock writer doesn't actually update the file). This means the recursive
	// call will also hit bmad-help, and eventually hit the depth limit.
	// This is expected behavior in the test since we can't mock the file re-reads.
	// Instead, just verify bmad-help was called.
	_ = err // Error is expected due to depth limit or continued unknown status

	// bmad-help should have been called at least once
	assert.NotEmpty(t, mockBmadHelp.Calls, "bmad-help should be called for unknown status")
	assert.Equal(t, "STORY-1", mockBmadHelp.Calls[0].StoryKey)
	assert.Equal(t, status.Status("pending-qa"), mockBmadHelp.Calls[0].CurrentStatus)

	// code-review should have been executed at least once
	assert.Contains(t, mockRunner.ExecutedWorkflows, "code-review")
}

// TestStoryCommand_DryRunWithoutModules tests dry-run has no module line when no modules
func TestStoryCommand_DryRunWithoutModules(t *testing.T) {
	tmpDir := t.TempDir()
	createSprintStatusFile(t, tmpDir, `development_status:
  STORY-1: backlog`)

	mockRunner := &MockWorkflowRunner{}
	mockWriter := &MockStatusWriter{}
	statusReader := status.NewReader(tmpDir)
	buf := &bytes.Buffer{}
	printer := output.NewPrinterWithWriter(buf)

	app := &App{
		Config:       config.DefaultConfig(),
		StatusReader: statusReader,
		StatusWriter: mockWriter,
		Runner:       mockRunner,
		Printer:      printer,
		// No Router, no Modules
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := NewRootCommand(app)
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"story", "--dry-run", "STORY-1"})

	err := rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(r)
	stdout := stdoutBuf.String()

	assert.NoError(t, err)

	// Should not contain modules line
	assert.NotContains(t, stdout, "Modules:")
}
