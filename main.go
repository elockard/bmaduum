package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Minimal structs - only what we need to extract
type StreamEvent struct {
	Type          string          `json:"type"`
	Subtype       string          `json:"subtype,omitempty"`
	Message       *MessageContent `json:"message,omitempty"`
	ToolUseResult *ToolResult     `json:"tool_use_result,omitempty"`
}

type MessageContent struct {
	Content []ContentBlock `json:"content,omitempty"`
}

type ContentBlock struct {
	Type  string     `json:"type"`
	Text  string     `json:"text,omitempty"`
	Name  string     `json:"name,omitempty"`
	Input *ToolInput `json:"input,omitempty"`
}

type ToolInput struct {
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Content     string `json:"content,omitempty"`
}

type ToolResult struct {
	Stdout      string `json:"stdout,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	Interrupted bool   `json:"interrupted,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-story":
		if len(os.Args) < 3 {
			fmt.Println("Error: create-story requires a story key")
			fmt.Println("Usage: bmad-automate create-story <story-key>")
			os.Exit(1)
		}
		storyKey := os.Args[2]
		prompt := fmt.Sprintf("/bmad:bmm:workflows:create-story - Create story: %s. Do not ask questions.", storyKey)
		os.Exit(runClaude(prompt, fmt.Sprintf("create-story: %s", storyKey)))

	case "dev-story":
		if len(os.Args) < 3 {
			fmt.Println("Error: dev-story requires a story key")
			fmt.Println("Usage: bmad-automate dev-story <story-key>")
			os.Exit(1)
		}
		storyKey := os.Args[2]
		prompt := fmt.Sprintf("/bmad:bmm:workflows:dev-story - Work on story: %s. Complete all tasks. Run tests after each implementation. Do not ask clarifying questions - use best judgment based on existing patterns.", storyKey)
		os.Exit(runClaude(prompt, fmt.Sprintf("dev-story: %s", storyKey)))

	case "code-review":
		if len(os.Args) < 3 {
			fmt.Println("Error: code-review requires a story key")
			fmt.Println("Usage: bmad-automate code-review <story-key>")
			os.Exit(1)
		}
		storyKey := os.Args[2]
		prompt := fmt.Sprintf("/bmad:bmm:workflows:code-review - Review story: %s. When presenting fix options, always choose to auto-fix all issues immediately. Do not wait for user input.", storyKey)
		os.Exit(runClaude(prompt, fmt.Sprintf("code-review: %s", storyKey)))

	case "git-commit":
		if len(os.Args) < 3 {
			fmt.Println("Error: git-commit requires a story key")
			fmt.Println("Usage: bmad-automate git-commit <story-key>")
			os.Exit(1)
		}
		storyKey := os.Args[2]
		prompt := fmt.Sprintf("Commit all changes for story %s with a descriptive commit message following conventional commits format. Then push to the current branch. Do not ask questions.", storyKey)
		os.Exit(runClaude(prompt, fmt.Sprintf("git-commit: %s", storyKey)))

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: run requires a story key")
			fmt.Println("Usage: bmad-automate run <story-key>")
			os.Exit(1)
		}
		storyKey := os.Args[2]
		os.Exit(runFullCycle(storyKey))

	case "queue":
		if len(os.Args) < 3 {
			fmt.Println("Error: queue requires at least one story key")
			fmt.Println("Usage: bmad-automate queue <story-key> [story-key...]")
			os.Exit(1)
		}
		storyKeys := os.Args[2:]
		os.Exit(runQueue(storyKeys))

	case "raw":
		// Raw mode - pass prompt directly (for testing)
		if len(os.Args) < 3 {
			fmt.Println("Error: raw requires a prompt")
			fmt.Println("Usage: bmad-automate raw \"your prompt\"")
			os.Exit(1)
		}
		prompt := strings.Join(os.Args[2:], " ")
		os.Exit(runClaude(prompt, "raw"))

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BMAD Automation")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  bmad-automate queue <story-key>...       Run full cycle on multiple stories")
	fmt.Println("  bmad-automate run <story-key>            Run full cycle (create → dev → review → commit)")
	fmt.Println("  bmad-automate create-story <story-key>   Run create-story workflow")
	fmt.Println("  bmad-automate dev-story <story-key>      Run dev-story workflow")
	fmt.Println("  bmad-automate code-review <story-key>    Run code-review workflow")
	fmt.Println("  bmad-automate git-commit <story-key>     Commit and push changes")
	fmt.Println("  bmad-automate raw \"<prompt>\"             Run arbitrary prompt")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  bmad-automate queue 6-5 6-6 6-7 6-8")
	fmt.Println("  bmad-automate run 6-4-fee-rebalancing")
	fmt.Println("  bmad-automate create-story 6-4-fee-rebalancing")
}

type step struct {
	name   string
	prompt string
}

func runFullCycle(storyKey string) int {
	totalStart := time.Now()

	steps := []step{
		{
			name:   "create-story",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:create-story - Create story: %s. Do not ask questions.", storyKey),
		},
		{
			name:   "dev-story",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:dev-story - Work on story: %s. Complete all tasks. Run tests after each implementation. Do not ask clarifying questions - use best judgment based on existing patterns.", storyKey),
		},
		{
			name:   "code-review",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:code-review - Review story: %s. When presenting fix options, always choose to auto-fix all issues immediately. Do not wait for user input.", storyKey),
		},
		{
			name:   "git-commit",
			prompt: fmt.Sprintf("Commit all changes for story %s with a descriptive commit message following conventional commits format. Then push to the current branch. Do not ask questions.", storyKey),
		},
	}

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  BMAD Full Cycle: %s\n", storyKey)
	fmt.Printf("║  Steps: create-story → dev-story → code-review → git-commit\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	durations := make([]time.Duration, len(steps))

	for i, s := range steps {
		fmt.Printf("┌─────────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│  [%d/%d] %s\n", i+1, len(steps), s.name)
		fmt.Printf("└─────────────────────────────────────────────────────────────────┘\n")

		stepStart := time.Now()
		exitCode := runClaude(s.prompt, fmt.Sprintf("%s: %s", s.name, storyKey))
		durations[i] = time.Since(stepStart)

		if exitCode != 0 {
			fmt.Printf("\n")
			fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
			fmt.Printf("║  ✗ CYCLE FAILED at step: %s\n", s.name)
			fmt.Printf("║  Story: %s\n", storyKey)
			fmt.Printf("║  Duration: %s\n", time.Since(totalStart).Round(time.Millisecond))
			fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")
			return exitCode
		}

		fmt.Printf("\n")
	}

	totalDuration := time.Since(totalStart)

	fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  ✓ CYCLE COMPLETE\n")
	fmt.Printf("║  Story: %s\n", storyKey)
	fmt.Printf("╠═══════════════════════════════════════════════════════════════╣\n")
	for i, s := range steps {
		fmt.Printf("║  [%d] %-15s %s\n", i+1, s.name, durations[i].Round(time.Millisecond))
	}
	fmt.Printf("╠═══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Total: %s\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")

	return 0
}

type storyResult struct {
	key      string
	success  bool
	duration time.Duration
	failedAt string
}

func runQueue(storyKeys []string) int {
	queueStart := time.Now()
	results := make([]storyResult, 0, len(storyKeys))

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  BMAD Queue: %d stories\n", len(storyKeys))
	fmt.Printf("║  Stories: %s\n", truncate(strings.Join(storyKeys, ", "), 50))
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	for i, storyKey := range storyKeys {
		fmt.Printf("╭─────────────────────────────────────────────────────────────────╮\n")
		fmt.Printf("│  QUEUE [%d/%d]: %s\n", i+1, len(storyKeys), storyKey)
		fmt.Printf("╰─────────────────────────────────────────────────────────────────╯\n")

		storyStart := time.Now()
		exitCode := runFullCycleInternal(storyKey)
		duration := time.Since(storyStart)

		result := storyResult{
			key:      storyKey,
			success:  exitCode == 0,
			duration: duration,
		}

		if exitCode != 0 {
			result.failedAt = "cycle"
			results = append(results, result)

			// Print partial summary and exit
			printQueueSummary(results, storyKeys, queueStart)
			return exitCode
		}

		results = append(results, result)
		fmt.Printf("\n")
	}

	printQueueSummary(results, storyKeys, queueStart)
	return 0
}

func printQueueSummary(results []storyResult, allKeys []string, startTime time.Time) {
	totalDuration := time.Since(startTime)
	completed := 0
	failed := 0
	for _, r := range results {
		if r.success {
			completed++
		} else {
			failed++
		}
	}
	remaining := len(allKeys) - len(results)

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
	if failed == 0 && remaining == 0 {
		fmt.Printf("║  ✓ QUEUE COMPLETE\n")
	} else {
		fmt.Printf("║  ✗ QUEUE STOPPED\n")
	}
	fmt.Printf("╠═══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Completed: %d | Failed: %d | Remaining: %d\n", completed, failed, remaining)
	fmt.Printf("╠═══════════════════════════════════════════════════════════════╣\n")
	for _, r := range results {
		status := "✓"
		if !r.success {
			status = "✗"
		}
		fmt.Printf("║  %s %-30s %s\n", status, r.key, r.duration.Round(time.Second))
	}
	if remaining > 0 {
		for i := len(results); i < len(allKeys); i++ {
			fmt.Printf("║  ○ %-30s (skipped)\n", allKeys[i])
		}
	}
	fmt.Printf("╠═══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Total: %s\n", totalDuration.Round(time.Second))
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")
}

// runFullCycleInternal is like runFullCycle but returns exit code instead of printing final box
func runFullCycleInternal(storyKey string) int {
	totalStart := time.Now()

	steps := []step{
		{
			name:   "create-story",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:create-story - Create story: %s. Do not ask questions.", storyKey),
		},
		{
			name:   "dev-story",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:dev-story - Work on story: %s. Complete all tasks. Run tests after each implementation. Do not ask clarifying questions - use best judgment based on existing patterns.", storyKey),
		},
		{
			name:   "code-review",
			prompt: fmt.Sprintf("/bmad:bmm:workflows:code-review - Review story: %s. When presenting fix options, always choose to auto-fix all issues immediately. Do not wait for user input.", storyKey),
		},
		{
			name:   "git-commit",
			prompt: fmt.Sprintf("Commit all changes for story %s with a descriptive commit message following conventional commits format. Then push to the current branch. Do not ask questions.", storyKey),
		},
	}

	durations := make([]time.Duration, len(steps))

	for i, s := range steps {
		fmt.Printf("  [%d/%d] %s\n", i+1, len(steps), s.name)

		stepStart := time.Now()
		exitCode := runClaude(s.prompt, fmt.Sprintf("%s: %s", s.name, storyKey))
		durations[i] = time.Since(stepStart)

		if exitCode != 0 {
			fmt.Printf("  ✗ Failed at %s\n", s.name)
			return exitCode
		}
	}

	totalDuration := time.Since(totalStart)
	fmt.Printf("  ✓ Story complete in %s\n", totalDuration.Round(time.Second))

	return 0
}

func runClaude(prompt string, label string) int {
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("  Command: %s\n", label)
	fmt.Printf("  Prompt:  %s\n", truncate(prompt, 60))
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")

	startTime := time.Now()

	cmd := exec.Command("claude",
		"--dangerously-skip-permissions",
		"-p", prompt,
		"--output-format", "stream-json",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdout pipe: %v\n", err)
		return 1
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stderr pipe: %v\n", err)
		return 1
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting claude: %v\n", err)
		return 1
	}

	// Handle stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "[stderr] %s\n", scanner.Text())
		}
	}()

	// Process streaming JSON from stdout
	scanner := bufio.NewScanner(stdout)

	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		var event StreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip unparseable lines
			continue
		}

		switch event.Type {
		case "system":
			if event.Subtype == "init" {
				fmt.Printf("● Session started\n\n")
			}

		case "assistant":
			if event.Message != nil {
				for _, block := range event.Message.Content {
					switch block.Type {
					case "text":
						if block.Text != "" {
							fmt.Printf("Claude: %s\n\n", block.Text)
						}
					case "tool_use":
						printToolUse(block)
					}
				}
			}

		case "user":
			// Tool results
			if event.ToolUseResult != nil {
				printToolResult(event.ToolUseResult)
			}

		case "result":
			// Final result - session complete
			fmt.Printf("● Session complete\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdout: %v\n", err)
	}

	// Wait for command to finish and get exit code
	err = cmd.Wait()

	duration := time.Since(startTime)
	exitCode := 0

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════════\n")
	if exitCode == 0 {
		fmt.Printf("  ✓ SUCCESS | Duration: %s\n", duration.Round(time.Millisecond))
	} else {
		fmt.Printf("  ✗ FAILED  | Duration: %s | Exit code: %d\n", duration.Round(time.Millisecond), exitCode)
	}
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")

	return exitCode
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printToolUse(block ContentBlock) {
	fmt.Printf("┌─ Tool: %s\n", block.Name)

	if block.Input != nil {
		if block.Input.Description != "" {
			fmt.Printf("│  %s\n", block.Input.Description)
		}
		if block.Input.Command != "" {
			fmt.Printf("│  $ %s\n", block.Input.Command)
		}
		if block.Input.FilePath != "" {
			fmt.Printf("│  File: %s\n", block.Input.FilePath)
		}
	}

	fmt.Printf("└─\n")
}

func printToolResult(result *ToolResult) {
	if result.Stdout != "" {
		// Truncate long output
		output := result.Stdout
		lines := strings.Split(output, "\n")
		if len(lines) > 20 {
			output = strings.Join(lines[:10], "\n") +
				fmt.Sprintf("\n  ... (%d lines omitted) ...\n", len(lines)-20) +
				strings.Join(lines[len(lines)-10:], "\n")
		}
		fmt.Printf("   %s\n\n", strings.ReplaceAll(output, "\n", "\n   "))
	}
	if result.Stderr != "" {
		fmt.Printf("   [stderr] %s\n\n", result.Stderr)
	}
}
