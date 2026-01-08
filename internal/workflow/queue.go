package workflow

import (
	"context"
	"fmt"
	"time"

	"bmad-automate/internal/output"
)

// QueueRunner processes multiple stories in sequence.
type QueueRunner struct {
	runner *Runner
}

// NewQueueRunner creates a new queue runner.
func NewQueueRunner(runner *Runner) *QueueRunner {
	return &QueueRunner{runner: runner}
}

// RunQueue executes the full cycle for each story in the queue.
// It stops on the first failure.
func (q *QueueRunner) RunQueue(ctx context.Context, storyKeys []string) int {
	queueStart := time.Now()
	results := make([]output.StoryResult, 0, len(storyKeys))

	q.runner.printer.QueueHeader(len(storyKeys), storyKeys)

	for i, storyKey := range storyKeys {
		q.runner.printer.QueueStoryStart(i+1, len(storyKeys), storyKey)

		storyStart := time.Now()
		exitCode := q.runFullCycleInternal(ctx, storyKey)
		duration := time.Since(storyStart)

		result := output.StoryResult{
			Key:      storyKey,
			Success:  exitCode == 0,
			Duration: duration,
		}

		if exitCode != 0 {
			result.FailedAt = "cycle"
			results = append(results, result)

			// Print partial summary and exit
			q.runner.printer.QueueSummary(results, storyKeys, time.Since(queueStart))
			return exitCode
		}

		results = append(results, result)
		fmt.Println() // Add spacing between stories
	}

	q.runner.printer.QueueSummary(results, storyKeys, time.Since(queueStart))
	return 0
}

// runFullCycleInternal runs a full cycle without the outer summary box.
// Used by queue to avoid duplicate boxing.
func (q *QueueRunner) runFullCycleInternal(ctx context.Context, storyKey string) int {
	totalStart := time.Now()

	// Build steps from config
	stepNames := q.runner.config.GetFullCycleSteps()
	steps := make([]Step, 0, len(stepNames))

	for _, name := range stepNames {
		prompt, err := q.runner.config.GetPrompt(name, storyKey)
		if err != nil {
			fmt.Printf("  Error building step %s: %v\n", name, err)
			return 1
		}
		steps = append(steps, Step{Name: name, Prompt: prompt})
	}

	for i, step := range steps {
		fmt.Printf("  [%d/%d] %s\n", i+1, len(steps), step.Name)

		stepStart := time.Now()
		exitCode := q.runner.runClaude(ctx, step.Prompt, fmt.Sprintf("%s: %s", step.Name, storyKey))
		duration := time.Since(stepStart)

		if exitCode != 0 {
			fmt.Printf("  ✗ Failed at %s\n", step.Name)
			return exitCode
		}

		_ = duration // Duration tracking for potential future use
	}

	totalDuration := time.Since(totalStart)
	fmt.Printf("  ✓ Story complete in %s\n", totalDuration.Round(time.Second))

	return 0
}
