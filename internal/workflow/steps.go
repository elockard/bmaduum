// Package workflow provides workflow orchestration for bmad-automate.
//
// Workflows execute Claude CLI with configured prompts to automate development tasks.
// The package supports single workflow execution, raw prompt execution, and batch
// processing of multiple stories through a queue system.
//
// Key types:
//   - [Runner] orchestrates individual Claude executions with output formatting
//   - [QueueRunner] processes multiple stories in sequence with status-based routing
//   - [Step] defines a workflow step with its name and expanded prompt
//
// The [Runner] requires a [claude.Executor] for spawning Claude CLI processes.
// Workflow prompts are configured in the config package and support Go template
// expansion with story keys.
package workflow

import "time"

// Step represents a single step in a workflow execution.
//
// Name identifies the workflow (e.g., "analyze", "implement", "test") and is used
// for display and configuration lookup. Prompt contains the fully-expanded prompt
// text after template substitution with the story key.
type Step struct {
	// Name is the workflow name used for display and configuration lookup.
	Name string
	// Prompt is the expanded prompt text to send to Claude CLI.
	Prompt string
}

// StepResult captures the outcome of executing a single workflow step.
//
// This type tracks execution metadata including timing, exit codes, and success
// status. It is used by [Runner.RunFullCycle] to aggregate results for the
// cycle summary display.
type StepResult struct {
	// Name is the workflow name that was executed.
	Name string
	// Duration is the wall-clock time spent executing the step.
	Duration time.Duration
	// ExitCode is the exit code returned by Claude CLI (0 = success).
	ExitCode int
	// Success indicates whether the step completed successfully.
	Success bool
}

// IsSuccess returns true if the step completed successfully.
//
// This is a convenience method that checks if ExitCode equals 0.
func (r StepResult) IsSuccess() bool {
	return r.ExitCode == 0
}
