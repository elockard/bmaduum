package workflow

import (
	"context"
	"fmt"
	"time"

	"bmad-automate/internal/claude"
	"bmad-automate/internal/config"
	"bmad-automate/internal/output"
)

// Runner executes workflows using Claude.
type Runner struct {
	executor claude.Executor
	printer  output.Printer
	config   *config.Config
}

// NewRunner creates a new workflow runner.
func NewRunner(executor claude.Executor, printer output.Printer, cfg *config.Config) *Runner {
	return &Runner{
		executor: executor,
		printer:  printer,
		config:   cfg,
	}
}

// RunSingle executes a single workflow step.
func (r *Runner) RunSingle(ctx context.Context, workflowName, storyKey string) int {
	prompt, err := r.config.GetPrompt(workflowName, storyKey)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	label := fmt.Sprintf("%s: %s", workflowName, storyKey)
	return r.runClaude(ctx, prompt, label)
}

// RunRaw executes an arbitrary prompt.
func (r *Runner) RunRaw(ctx context.Context, prompt string) int {
	return r.runClaude(ctx, prompt, "raw")
}

// RunFullCycle executes all steps in the full cycle workflow.
func (r *Runner) RunFullCycle(ctx context.Context, storyKey string) int {
	totalStart := time.Now()

	// Build steps from config
	stepNames := r.config.GetFullCycleSteps()
	steps := make([]Step, 0, len(stepNames))

	for _, name := range stepNames {
		prompt, err := r.config.GetPrompt(name, storyKey)
		if err != nil {
			fmt.Printf("Error building step %s: %v\n", name, err)
			return 1
		}
		steps = append(steps, Step{Name: name, Prompt: prompt})
	}

	r.printer.CycleHeader(storyKey)

	results := make([]output.StepResult, len(steps))

	for i, step := range steps {
		r.printer.StepStart(i+1, len(steps), step.Name)

		stepStart := time.Now()
		exitCode := r.runClaude(ctx, step.Prompt, fmt.Sprintf("%s: %s", step.Name, storyKey))
		duration := time.Since(stepStart)

		results[i] = output.StepResult{
			Name:     step.Name,
			Duration: duration,
			Success:  exitCode == 0,
		}

		if exitCode != 0 {
			r.printer.CycleFailed(storyKey, step.Name, time.Since(totalStart))
			return exitCode
		}

		fmt.Println() // Add spacing between steps
	}

	r.printer.CycleSummary(storyKey, results, time.Since(totalStart))
	return 0
}

// runClaude executes Claude with the given prompt and handles output.
func (r *Runner) runClaude(ctx context.Context, prompt, label string) int {
	r.printer.CommandHeader(label, prompt, r.config.Output.TruncateLength)

	startTime := time.Now()

	handler := func(event claude.Event) {
		r.handleEvent(event)
	}

	exitCode, err := r.executor.ExecuteWithResult(ctx, prompt, handler)
	if err != nil {
		fmt.Printf("Error executing claude: %v\n", err)
		exitCode = 1
	}

	duration := time.Since(startTime)
	r.printer.CommandFooter(duration, exitCode == 0, exitCode)

	return exitCode
}

// handleEvent processes a single event from Claude.
func (r *Runner) handleEvent(event claude.Event) {
	switch {
	case event.SessionStarted:
		r.printer.SessionStart()

	case event.IsText():
		r.printer.Text(event.Text)

	case event.IsToolUse():
		r.printer.ToolUse(event.ToolName, event.ToolDescription, event.ToolCommand, event.ToolFilePath)

	case event.IsToolResult():
		r.printer.ToolResult(event.ToolStdout, event.ToolStderr, r.config.Output.TruncateLines)

	case event.SessionComplete:
		r.printer.SessionEnd(0, true) // Duration handled elsewhere
	}
}
