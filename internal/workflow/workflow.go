package workflow

import (
	"context"
	"fmt"
	"os"
	"time"

	"bmaduum/internal/claude"
	"bmaduum/internal/config"
	"bmaduum/internal/output/core"
	"bmaduum/internal/output/progress"
	"bmaduum/internal/ratelimit"
)

// Runner orchestrates workflow execution using Claude CLI.
//
// Runner is the primary executor for development workflows. It combines a
// [claude.Executor] for spawning Claude processes, a [core.Printer] for
// formatted terminal output, and a [config.Config] for prompt templates.
//
// Use [NewRunner] to create a properly initialized Runner instance.
type Runner struct {
	executor   claude.Executor
	printer    core.Printer
	progress   *progress.Line
	config     *config.Config
	detector   *ratelimit.Detector
	correlator *ToolCorrelator // Correlates tool uses with their results
}

// NewRunner creates a new workflow runner with the specified dependencies.
//
// Parameters:
//   - executor: The [claude.Executor] implementation for running Claude CLI
//   - printer: The [core.Printer] for formatted terminal output
//   - cfg: The configuration containing workflow prompt templates
//
// The executor typically uses [claude.NewExecutor] in production or
// [claude.MockExecutor] for testing.
func NewRunner(executor claude.Executor, printer core.Printer, cfg *config.Config) *Runner {
	return &Runner{
		executor:   executor,
		printer:    printer,
		progress:   progress.NewLine(os.Stdout),
		config:     cfg,
		detector:   ratelimit.NewDetector(),
		correlator: NewToolCorrelator(),
	}
}

// SetOperation sets the operation context for display in the status bar.
// This is typically called by CLI commands to show the broader context
// (e.g., "Epic 6", "Story 2/3").
func (r *Runner) SetOperation(operation string) {
	r.progress.SetOperation(operation)
}

// RunSingle executes a single named workflow for a story.
//
// The workflowName must match a workflow defined in the configuration (e.g.,
// "analyze", "implement", "test"). The storyKey is substituted into the
// workflow's prompt template.
//
// Returns the exit code from Claude CLI (0 for success, non-zero for failure).
func (r *Runner) RunSingle(ctx context.Context, workflowName, storyKey string) int {
	prompt, err := r.config.GetPrompt(workflowName, storyKey)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	label := fmt.Sprintf("%s: %s", workflowName, storyKey)
	model := r.config.GetModel(workflowName)
	return r.runClaude(ctx, prompt, label, model)
}

// RunRaw executes an arbitrary prompt without template expansion.
//
// Use this method for one-off or custom prompts that don't correspond to
// configured workflows. The prompt is passed directly to Claude CLI.
//
// Returns the exit code from Claude CLI (0 for success, non-zero for failure).
func (r *Runner) RunRaw(ctx context.Context, prompt string) int {
	return r.runClaude(ctx, prompt, "raw", "")
}

// runClaude executes Claude CLI with the given prompt and handles streaming output.
//
// This is the core execution method used by all public Runner methods.
// It displays a command header, streams events to the printer via handleEvent,
// updates the progress line, and displays a footer with timing and exit status.
func (r *Runner) runClaude(ctx context.Context, prompt, label, model string) int {
	// Reset correlator for new execution
	r.correlator.Reset()

	// Initialize progress line FIRST (sets up scroll region at bottom)
	// This must happen before any output so content flows naturally
	r.progress.Init()

	// Set initial step info
	r.progress.SetStepInfo(0, 0, label, "", model)

	// Now print header (it will scroll within the scroll region)
	r.printer.CommandHeader(label, prompt, r.config.Output.TruncateLength)

	startTime := time.Now()

	// Event handler that routes events and updates progress
	handler := func(event claude.Event) {
		// Track token usage - estimate from text if actual counts are 0
		if event.InputTokens > 0 || event.OutputTokens > 0 {
			r.progress.AddTokens(event.InputTokens, event.OutputTokens)
		} else if event.IsText() && len(event.Text) > 0 {
			// Estimate tokens: roughly 4 characters per token for English text
			// This is a rough approximation since Claude CLI doesn't provide streaming token counts
			estimatedTokens := (len(event.Text) + 3) / 4
			r.progress.AddTokens(0, estimatedTokens)
		}

		// Record first response for thinking time calculation
		if event.IsText() || event.IsToolUse() {
			r.progress.RecordFirstResponse()
		}

		// Update progress based on event type
		switch {
		case event.IsToolUse():
			r.progress.SetCurrentTool(event.ToolName)
			r.progress.IncrementToolCount()
		case event.IsToolResult():
			r.progress.SetCurrentTool("") // Back to thinking
		}

		// Print the event (output scrolls below status bar)
		r.handleEvent(event)

		// Check for rate limit in stderr after printing
		if event.IsToolResult() && event.ToolStderr != "" {
			info := r.detector.CheckLine(event.ToolStderr)
			if info.IsRateLimit {
				wait := r.detector.WaitTime(info)
				r.progress.UpdateWithRateLimit(0, 0, label, "", time.Since(startTime), info.ResetTime)
				time.Sleep(wait)
				r.progress.SetCurrentTool("")
			}
		}
	}

	exitCode, err := r.executor.ExecuteWithResult(ctx, prompt, handler, model)
	if err != nil {
		fmt.Printf("Error executing claude: %v\n", err)
		exitCode = 1
	}

	duration := time.Since(startTime)
	r.progress.Done(exitCode == 0, duration)
	r.printer.CommandFooter(duration, exitCode == 0, exitCode)

	return exitCode
}

// handleEvent routes a Claude streaming event to the appropriate printer method.
// Tool uses are buffered and correlated with their results to print them together,
// matching Claude Code's display behavior.
func (r *Runner) handleEvent(event claude.Event) {
	switch {
	case event.SessionStarted:
		r.printer.SessionStart()

	case event.IsText():
		// Flush any pending tools before printing text
		r.flushPendingTools()
		r.printer.Text(event.Text)

	case event.IsToolUse():
		// Buffer tool use for correlation with its result
		params := EventToToolParams(event)
		r.correlator.AddToolUse(event.ToolID, params)

	case event.IsToolResult():
		// Match result with pending tool use and print together
		if params, found := r.correlator.MatchResult(event.ToolUseID); found {
			r.printer.ToolUse(params)
			r.printer.ToolResult(event.ToolStdout, event.ToolStderr, r.config.Output.TruncateLines)
		} else {
			// No matching tool use found, just print the result
			r.printer.ToolResult(event.ToolStdout, event.ToolStderr, r.config.Output.TruncateLines)
		}

	case event.SessionComplete:
		// Flush any remaining pending tools
		r.flushPendingTools()
		r.printer.SessionEnd(0, true)
	}
}

// flushPendingTools prints any buffered tool uses without waiting for results.
// This is called when text arrives or the session ends.
func (r *Runner) flushPendingTools() {
	for _, tool := range r.correlator.Flush() {
		r.printer.ToolUse(tool.Params)
	}
}
