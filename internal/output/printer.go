package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// StepResult represents the result of a single step execution.
type StepResult struct {
	Name     string
	Duration time.Duration
	Success  bool
}

// StoryResult represents the result of processing a story in a queue.
type StoryResult struct {
	Key      string
	Success  bool
	Duration time.Duration
	FailedAt string
}

// Printer defines the interface for terminal output operations.
type Printer interface {
	// Session lifecycle
	SessionStart()
	SessionEnd(duration time.Duration, success bool)

	// Step progress
	StepStart(step, total int, name string)
	StepEnd(duration time.Duration, success bool)

	// Tool usage
	ToolUse(name, description, command, filePath string)
	ToolResult(stdout, stderr string, truncateLines int)

	// Content
	Text(message string)
	Divider()

	// Full cycle output
	CycleHeader(storyKey string)
	CycleSummary(storyKey string, steps []StepResult, totalDuration time.Duration)
	CycleFailed(storyKey string, failedStep string, duration time.Duration)

	// Queue output
	QueueHeader(count int, stories []string)
	QueueStoryStart(index, total int, storyKey string)
	QueueSummary(results []StoryResult, allKeys []string, totalDuration time.Duration)

	// Command info
	CommandHeader(label, prompt string, truncateLength int)
	CommandFooter(duration time.Duration, success bool, exitCode int)
}

// DefaultPrinter implements Printer with lipgloss styling.
type DefaultPrinter struct {
	out io.Writer
}

// NewPrinter creates a new DefaultPrinter that writes to stdout.
func NewPrinter() *DefaultPrinter {
	return &DefaultPrinter{out: os.Stdout}
}

// NewPrinterWithWriter creates a new DefaultPrinter with a custom writer.
func NewPrinterWithWriter(w io.Writer) *DefaultPrinter {
	return &DefaultPrinter{out: w}
}

func (p *DefaultPrinter) writeln(format string, args ...interface{}) {
	fmt.Fprintf(p.out, format+"\n", args...)
}

// SessionStart prints session start indicator.
func (p *DefaultPrinter) SessionStart() {
	p.writeln("%s Session started\n", iconInProgress)
}

// SessionEnd prints session end with status.
func (p *DefaultPrinter) SessionEnd(duration time.Duration, success bool) {
	p.writeln("%s Session complete", iconInProgress)
}

// StepStart prints step start header.
func (p *DefaultPrinter) StepStart(step, total int, name string) {
	header := fmt.Sprintf("[%d/%d] %s", step, total, name)
	p.writeln(stepHeaderStyle.Render(header))
}

// StepEnd prints step completion status.
func (p *DefaultPrinter) StepEnd(duration time.Duration, success bool) {
	// Step end is usually handled by CommandFooter
}

// ToolUse prints tool invocation details.
func (p *DefaultPrinter) ToolUse(name, description, command, filePath string) {
	p.writeln("%s Tool: %s", iconTool, toolNameStyle.Render(name))

	if description != "" {
		p.writeln("%s  %s", iconToolLine, description)
	}
	if command != "" {
		p.writeln("%s  $ %s", iconToolLine, command)
	}
	if filePath != "" {
		p.writeln("%s  File: %s", iconToolLine, filePath)
	}

	p.writeln(iconToolEnd)
}

// ToolResult prints tool execution results.
func (p *DefaultPrinter) ToolResult(stdout, stderr string, truncateLines int) {
	if stdout != "" {
		output := truncateOutput(stdout, truncateLines)
		// Indent each line
		indented := "   " + strings.ReplaceAll(output, "\n", "\n   ")
		p.writeln("%s\n", indented)
	}
	if stderr != "" {
		p.writeln("   %s\n", mutedStyle.Render("[stderr] "+stderr))
	}
}

// Text prints a text message from Claude.
func (p *DefaultPrinter) Text(message string) {
	if message != "" {
		p.writeln("Claude: %s\n", message)
	}
}

// Divider prints a visual divider.
func (p *DefaultPrinter) Divider() {
	p.writeln(dividerStyle.Render(strings.Repeat("═", 65)))
}

// CycleHeader prints the header for a full cycle run.
func (p *DefaultPrinter) CycleHeader(storyKey string) {
	p.writeln("")
	content := fmt.Sprintf("BMAD Full Cycle: %s\nSteps: create-story → dev-story → code-review → git-commit", storyKey)
	p.writeln(headerStyle.Render(content))
	p.writeln("")
}

// CycleSummary prints the summary after a successful cycle.
func (p *DefaultPrinter) CycleSummary(storyKey string, steps []StepResult, totalDuration time.Duration) {
	var sb strings.Builder

	sb.WriteString(successStyle.Render(iconSuccess+" CYCLE COMPLETE") + "\n")
	sb.WriteString(fmt.Sprintf("Story: %s\n", storyKey))
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	for i, step := range steps {
		sb.WriteString(fmt.Sprintf("[%d] %-15s %s\n", i+1, step.Name, step.Duration.Round(time.Millisecond)))
	}

	sb.WriteString(strings.Repeat("─", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %s", totalDuration.Round(time.Millisecond)))

	p.writeln(summaryStyle.Render(sb.String()))
}

// CycleFailed prints failure information when a cycle fails.
func (p *DefaultPrinter) CycleFailed(storyKey string, failedStep string, duration time.Duration) {
	var sb strings.Builder

	sb.WriteString(errorStyle.Render(iconError+" CYCLE FAILED") + "\n")
	sb.WriteString(fmt.Sprintf("Story: %s\n", storyKey))
	sb.WriteString(fmt.Sprintf("Failed at: %s\n", failedStep))
	sb.WriteString(fmt.Sprintf("Duration: %s", duration.Round(time.Millisecond)))

	p.writeln(summaryStyle.Render(sb.String()))
}

// QueueHeader prints the header for a queue run.
func (p *DefaultPrinter) QueueHeader(count int, stories []string) {
	p.writeln("")
	storiesList := truncateString(strings.Join(stories, ", "), 50)
	content := fmt.Sprintf("BMAD Queue: %d stories\nStories: %s", count, storiesList)
	p.writeln(headerStyle.Render(content))
	p.writeln("")
}

// QueueStoryStart prints the header for starting a story in a queue.
func (p *DefaultPrinter) QueueStoryStart(index, total int, storyKey string) {
	header := fmt.Sprintf("QUEUE [%d/%d]: %s", index, total, storyKey)
	p.writeln(queueHeaderStyle.Render(header))
}

// QueueSummary prints the summary after a queue completes or fails.
func (p *DefaultPrinter) QueueSummary(results []StoryResult, allKeys []string, totalDuration time.Duration) {
	completed := 0
	failed := 0
	for _, r := range results {
		if r.Success {
			completed++
		} else {
			failed++
		}
	}
	remaining := len(allKeys) - len(results)

	var sb strings.Builder

	if failed == 0 && remaining == 0 {
		sb.WriteString(successStyle.Render(iconSuccess+" QUEUE COMPLETE") + "\n")
	} else {
		sb.WriteString(errorStyle.Render(iconError+" QUEUE STOPPED") + "\n")
	}

	sb.WriteString(strings.Repeat("─", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Completed: %d | Failed: %d | Remaining: %d\n", completed, failed, remaining))
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	for _, r := range results {
		status := successStyle.Render(iconSuccess)
		if !r.Success {
			status = errorStyle.Render(iconError)
		}
		sb.WriteString(fmt.Sprintf("%s %-30s %s\n", status, r.Key, r.Duration.Round(time.Second)))
	}

	if remaining > 0 {
		for i := len(results); i < len(allKeys); i++ {
			sb.WriteString(fmt.Sprintf("%s %-30s (skipped)\n", mutedStyle.Render(iconPending), allKeys[i]))
		}
	}

	sb.WriteString(strings.Repeat("─", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %s", totalDuration.Round(time.Second)))

	p.writeln(summaryStyle.Render(sb.String()))
}

// CommandHeader prints the header before running a command.
func (p *DefaultPrinter) CommandHeader(label, prompt string, truncateLength int) {
	p.Divider()
	p.writeln("  Command: %s", labelStyle.Render(label))
	p.writeln("  Prompt:  %s", truncateString(prompt, truncateLength))
	p.Divider()
	p.writeln("")
}

// CommandFooter prints the footer after a command completes.
func (p *DefaultPrinter) CommandFooter(duration time.Duration, success bool, exitCode int) {
	p.writeln("")
	p.Divider()
	if success {
		p.writeln("  %s | Duration: %s", successStyle.Render(iconSuccess+" SUCCESS"), duration.Round(time.Millisecond))
	} else {
		p.writeln("  %s | Duration: %s | Exit code: %d", errorStyle.Render(iconError+" FAILED"), duration.Round(time.Millisecond), exitCode)
	}
	p.Divider()
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// truncateOutput truncates output to maxLines, showing first and last portions.
func truncateOutput(output string, maxLines int) string {
	if maxLines <= 0 {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}

	half := maxLines / 2
	omitted := len(lines) - maxLines

	first := strings.Join(lines[:half], "\n")
	last := strings.Join(lines[len(lines)-half:], "\n")

	return fmt.Sprintf("%s\n  ... (%d lines omitted) ...\n%s", first, omitted, last)
}
