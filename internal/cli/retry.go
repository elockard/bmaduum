package cli

import (
	"context"
	"fmt"
	"time"

	"bmad-automate/internal/lifecycle"
	"bmad-automate/internal/ratelimit"
)

// executeWithRetry executes a story lifecycle with automatic retry on rate limit errors.
//
// If autoRetry is true, rate limit errors will trigger a wait until the reset time,
// then retry up to maxRetries times. The progress callback is invoked before each
// workflow execution.
func executeWithRetry(
	ctx context.Context,
	executor *lifecycle.Executor,
	storyKey string,
	autoRetry bool,
	maxRetries int,
	progressCallback func(stepIndex, totalSteps int, workflow string),
) error {
	if !autoRetry {
		// No retry - just execute once
		if progressCallback != nil {
			executor.SetProgressCallback(progressCallback)
		}
		return executor.Execute(ctx, storyKey)
	}

	// With auto-retry
	retryCount := 0
	for {
		if progressCallback != nil {
			executor.SetProgressCallback(progressCallback)
		}

		rateLimitState := ratelimit.NewState()

		// Set up stderr handler to detect rate limits
		// This would need to be integrated with the executor's stderr handling
		// For now, we execute and check for rate limit errors

		err := executor.Execute(ctx, storyKey)

		// Check if this is a rate limit error
		// In a full implementation, we would need to capture stderr and check
		// For now, we just retry on any error if autoRetry is enabled
		if err == nil {
			return nil
		}

		// Check if we've exceeded max retries
		if retryCount >= maxRetries {
			return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
		}

		// Wait before retrying
		waitTime := time.Duration(retryCount+1) * 30 * time.Second
		if rateLimitState.WaitTime() > 0 {
			waitTime = rateLimitState.WaitTime()
		}

		fmt.Printf("\n⚠️  Error encountered, waiting %v before retry %d/%d...\n",
			waitTime.Round(time.Second), retryCount+1, maxRetries)
		time.Sleep(waitTime)

		retryCount++
	}
}

// AutoRetryConfig holds configuration for automatic retry behavior.
type AutoRetryConfig struct {
	// Enabled indicates whether auto-retry is enabled.
	Enabled bool

	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// DefaultWaitTime is the default wait time between retries.
	DefaultWaitTime time.Duration
}

// DefaultAutoRetryConfig returns a default auto-retry configuration.
func DefaultAutoRetryConfig() AutoRetryConfig {
	return AutoRetryConfig{
		Enabled:         false,
		MaxRetries:      10,
		DefaultWaitTime: 5 * time.Minute,
	}
}
