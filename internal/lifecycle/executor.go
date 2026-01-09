// Package lifecycle orchestrates story lifecycle execution from current status to done.
//
// The lifecycle package provides [Executor] which runs stories through their complete
// workflow sequence (create->dev->review->commit) based on current status. Each step
// updates the story status automatically after successful completion.
//
// Key concepts:
//   - Lifecycle steps are determined by [router.GetLifecycle] based on current status
//   - Each step runs a workflow then updates status via [StatusWriter]
//   - Progress can be tracked via [ProgressCallback]
package lifecycle

import (
	"context"
	"fmt"

	"bmad-automate/internal/router"
	"bmad-automate/internal/status"
)

// WorkflowRunner is the interface for running workflows.
type WorkflowRunner interface {
	RunSingle(ctx context.Context, workflowName, storyKey string) int
}

// StatusReader is the interface for reading story status.
type StatusReader interface {
	GetStoryStatus(storyKey string) (status.Status, error)
}

// StatusWriter is the interface for updating story status.
type StatusWriter interface {
	UpdateStatus(storyKey string, newStatus status.Status) error
}

// ProgressCallback is called before each workflow step executes.
type ProgressCallback func(stepIndex, totalSteps int, workflow string)

// Executor runs the complete lifecycle for a story.
type Executor struct {
	runner           WorkflowRunner
	statusReader     StatusReader
	statusWriter     StatusWriter
	progressCallback ProgressCallback
}

// NewExecutor creates a new Executor with the given dependencies.
func NewExecutor(runner WorkflowRunner, reader StatusReader, writer StatusWriter) *Executor {
	return &Executor{
		runner:       runner,
		statusReader: reader,
		statusWriter: writer,
	}
}

// SetProgressCallback sets an optional callback that is invoked before each workflow step.
func (e *Executor) SetProgressCallback(cb ProgressCallback) {
	e.progressCallback = cb
}

// Execute runs the complete lifecycle for a story from its current status to done.
func (e *Executor) Execute(ctx context.Context, storyKey string) error {
	// Get current story status
	currentStatus, err := e.statusReader.GetStoryStatus(storyKey)
	if err != nil {
		return err
	}

	// Get lifecycle steps from current status
	steps, err := router.GetLifecycle(currentStatus)
	if err != nil {
		return err // Returns router.ErrStoryComplete for done stories
	}

	// Get total steps count for progress reporting
	totalSteps := len(steps)

	// Execute each step in sequence
	for i, step := range steps {
		// Call progress callback if set
		if e.progressCallback != nil {
			e.progressCallback(i+1, totalSteps, step.Workflow)
		}

		// Run the workflow
		exitCode := e.runner.RunSingle(ctx, step.Workflow, storyKey)
		if exitCode != 0 {
			return fmt.Errorf("workflow failed: %s returned exit code %d", step.Workflow, exitCode)
		}

		// Update status after successful workflow
		if err := e.statusWriter.UpdateStatus(storyKey, step.NextStatus); err != nil {
			return err
		}
	}

	return nil
}

// GetSteps returns the lifecycle steps for a story without executing them.
// This is used for dry-run preview to show what workflows would execute.
func (e *Executor) GetSteps(storyKey string) ([]router.LifecycleStep, error) {
	// Get current story status
	currentStatus, err := e.statusReader.GetStoryStatus(storyKey)
	if err != nil {
		return nil, err
	}

	// Get lifecycle steps from current status
	steps, err := router.GetLifecycle(currentStatus)
	if err != nil {
		return nil, err // Returns router.ErrStoryComplete for done stories
	}

	return steps, nil
}
