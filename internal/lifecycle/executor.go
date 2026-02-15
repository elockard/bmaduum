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
	"errors"
	"fmt"

	"bmaduum/internal/router"
	"bmaduum/internal/status"
)

// maxBmadHelpDepth limits recursive Execute calls when bmad-help resolves
// unknown statuses. This prevents infinite loops if bmad-help repeatedly
// returns statuses that the router doesn't recognize.
const maxBmadHelpDepth = 3

// WorkflowRunner is the interface for executing individual workflows.
//
// RunSingle executes a named workflow for a story and returns the exit code.
// An exit code of 0 indicates success; any non-zero value indicates failure.
// The [workflow.Runner] type implements this interface.
type WorkflowRunner interface {
	RunSingle(ctx context.Context, workflowName, storyKey string) int
}

// StatusReader is the interface for looking up story status.
//
// GetStoryStatus retrieves the current [status.Status] for a story key.
// It returns an error if the story cannot be found or the status file is invalid.
type StatusReader interface {
	GetStoryStatus(storyKey string) (status.Status, error)
}

// StatusWriter is the interface for persisting story status updates.
//
// UpdateStatus sets a new [status.Status] for a story after successful workflow completion.
// It returns an error if the status file cannot be written.
type StatusWriter interface {
	UpdateStatus(storyKey string, newStatus status.Status) error
}

// BmadHelpFallback resolves unknown statuses by invoking /bmad-help via Claude CLI.
//
// This is a last-resort fallback used when the standard router returns
// [router.ErrUnknownStatus]. Implementations should invoke /bmad-help and
// parse the response to determine the next workflow to execute.
//
// See the bmadhelp package for the production implementation.
type BmadHelpFallback interface {
	// ResolveWorkflow determines the next workflow for a story with an unknown status.
	// Returns the workflow name and expected next status.
	ResolveWorkflow(ctx context.Context, storyKey string, currentStatus status.Status) (workflow string, nextStatus status.Status, err error)
}

// ProgressCallback is invoked before each workflow step begins execution.
//
// The callback receives stepIndex (1-based), totalSteps count, and the workflow name.
// This enables progress reporting in the UI. The callback is optional and can be set
// via [Executor.SetProgressCallback].
type ProgressCallback func(stepIndex, totalSteps int, workflow string)

// Executor orchestrates the complete story lifecycle from current status to done.
//
// Executor uses dependency injection for testability: [WorkflowRunner] executes workflows,
// [StatusReader] looks up current status, and [StatusWriter] persists status updates.
// Use [NewExecutor] to create an instance and [Execute] to run the lifecycle.
//
// By default, the executor uses the hardcoded router for status-to-workflow mapping.
// Call [SetRouter] to use a manifest-driven router instead.
type Executor struct {
	runner           WorkflowRunner
	statusReader     StatusReader
	statusWriter     StatusWriter
	progressCallback ProgressCallback
	router           *router.Router
	bmadHelp         BmadHelpFallback
}

// NewExecutor creates a new Executor with the required dependencies.
//
// The runner executes workflows, reader looks up story status, and writer persists
// status updates. Progress callback is not set by default; use [SetProgressCallback]
// to enable progress reporting.
func NewExecutor(runner WorkflowRunner, reader StatusReader, writer StatusWriter) *Executor {
	return &Executor{
		runner:       runner,
		statusReader: reader,
		statusWriter: writer,
	}
}

// SetRouter configures a custom [router.Router] for status-to-workflow mapping.
//
// When set, the executor uses the provided router instead of the default hardcoded
// routing. This enables manifest-driven routing via [router.NewRouterFromManifest].
// If not set (or set to nil), the default package-level router functions are used.
func (e *Executor) SetRouter(r *router.Router) {
	e.router = r
}

// SetBmadHelp configures an optional bmad-help fallback for resolving unknown statuses.
//
// When set, the executor will invoke /bmad-help via Claude CLI as a last resort
// when the router returns [router.ErrUnknownStatus]. This enables handling of
// non-standard status values that aren't in the routing table.
//
// If not set (or set to nil), unknown statuses produce an immediate error.
func (e *Executor) SetBmadHelp(fb BmadHelpFallback) {
	e.bmadHelp = fb
}

// getLifecycle delegates to the configured router or falls back to the package-level function.
func (e *Executor) getLifecycle(s status.Status) ([]router.LifecycleStep, error) {
	if e.router != nil {
		return e.router.GetLifecycle(s)
	}
	return router.GetLifecycle(s)
}

// SetProgressCallback configures an optional progress callback for workflow execution.
//
// The callback receives the step index (1-based), total step count, and workflow name
// before each workflow begins. This is typically used to display progress information
// in the terminal UI.
func (e *Executor) SetProgressCallback(cb ProgressCallback) {
	e.progressCallback = cb
}

// Execute runs the complete story lifecycle from current status to done.
//
// Execute looks up the story's current status, determines the remaining workflow steps
// via [router.GetLifecycle], and runs each workflow in sequence. After each successful
// workflow, the story status is updated to the next state.
//
// When the router returns [router.ErrUnknownStatus] and a bmad-help fallback is
// configured (via [SetBmadHelp]), Execute invokes /bmad-help to get a single
// workflow recommendation. After executing that recommendation, it re-reads the
// status and continues with normal routing. This recursion is depth-limited to
// prevent infinite loops.
//
// Execute uses fail-fast behavior: it stops on the first error and returns immediately.
// Errors can occur from status lookup failure, workflow execution failure (non-zero exit),
// or status update failure. For stories already done, Execute returns [router.ErrStoryComplete].
func (e *Executor) Execute(ctx context.Context, storyKey string) error {
	return e.executeWithDepth(ctx, storyKey, 0)
}

// executeWithDepth is the internal implementation of Execute with depth tracking
// for bmad-help fallback recursion.
func (e *Executor) executeWithDepth(ctx context.Context, storyKey string, depth int) error {
	// Get current story status
	currentStatus, err := e.statusReader.GetStoryStatus(storyKey)
	if err != nil {
		return err
	}

	// Get lifecycle steps from current status
	steps, err := e.getLifecycle(currentStatus)
	usedBmadHelp := false
	if err != nil {
		if errors.Is(err, router.ErrUnknownStatus) && e.bmadHelp != nil {
			if depth >= maxBmadHelpDepth {
				return fmt.Errorf("unknown status %q: bmad-help fallback exceeded maximum depth (%d)", currentStatus, maxBmadHelpDepth)
			}

			workflow, nextStatus, helpErr := e.bmadHelp.ResolveWorkflow(ctx, storyKey, currentStatus)
			if helpErr != nil {
				return fmt.Errorf("unknown status %q and bmad-help fallback failed: %w", currentStatus, helpErr)
			}

			steps = []router.LifecycleStep{{
				Workflow:   workflow,
				NextStatus: nextStatus,
			}}
			usedBmadHelp = true
		} else {
			return err // Returns router.ErrStoryComplete for done stories, or ErrUnknownStatus without fallback
		}
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

	// If bmad-help bridged us from an unknown status, re-execute to continue
	// the lifecycle from the new (hopefully recognized) status.
	if usedBmadHelp {
		return e.executeWithDepth(ctx, storyKey, depth+1)
	}

	return nil
}

// GetSteps returns the remaining lifecycle steps for a story without executing them.
//
// GetSteps provides dry-run preview functionality, showing what workflows would execute
// and what status transitions would occur. This is useful for displaying the planned
// execution path before actually running workflows.
//
// Returns an error if status lookup fails. For stories already done, returns
// [router.ErrStoryComplete].
func (e *Executor) GetSteps(storyKey string) ([]router.LifecycleStep, error) {
	// Get current story status
	currentStatus, err := e.statusReader.GetStoryStatus(storyKey)
	if err != nil {
		return nil, err
	}

	// Get lifecycle steps from current status
	steps, err := e.getLifecycle(currentStatus)
	if err != nil {
		return nil, err // Returns router.ErrStoryComplete for done stories
	}

	return steps, nil
}
