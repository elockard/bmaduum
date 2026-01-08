// Package workflow provides workflow orchestration for bmad-automate.
package workflow

import "time"

// Step represents a single step in a workflow.
type Step struct {
	Name   string
	Prompt string
}

// StepResult represents the result of executing a step.
type StepResult struct {
	Name     string
	Duration time.Duration
	ExitCode int
	Success  bool
}

// IsSuccess returns true if the step completed successfully.
func (r StepResult) IsSuccess() bool {
	return r.ExitCode == 0
}
