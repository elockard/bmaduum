package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"bmaduum/internal/lifecycle"
	"bmaduum/internal/router"
)

func newEpicCommand(app *App) *cobra.Command {
	var dryRun bool
	var autoRetry bool

	cmd := &cobra.Command{
		Use:   "epic <epic-id> [epic-id...]",
		Short: "Run full lifecycle for all stories in one or more epics",
		Long: `Run the complete lifecycle for all stories in one or more epics to completion.

Finds all stories matching the pattern {epic-id}-{N}-* where N is numeric,
sorts them by story number, and runs each to completion before moving to the next.

For each story, executes all remaining workflows based on its current status:
  - backlog       → create-story → dev-story → code-review → git-commit → done
  - ready-for-dev → dev-story → code-review → git-commit → done
  - in-progress   → dev-story → code-review → git-commit → done
  - review        → code-review → git-commit → done
  - done          → skipped (story already complete)

The epic command stops on the first failure. Done stories are skipped and do not cause failure.
Status is updated in sprint-status.yaml after each successful workflow.

Use --dry-run to preview workflows without executing them.
Use --auto-retry to automatically retry on rate limit errors.

Example:
  bmaduum epic 6
  bmaduum epic 2 4 6`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			epicIDs := args

			// Create lifecycle executor with app dependencies
			executor := lifecycle.NewExecutor(app.Runner, app.StatusReader, app.StatusWriter)

			// Handle dry-run mode
			if dryRun {
				return runEpicDryRun(cmd, app, executor, epicIDs)
			}

			// Process each epic
			for i, epicID := range epicIDs {
				fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
				fmt.Printf("  Epic %d of %d: %s\n", i+1, len(epicIDs), epicID)
				fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")

				// Get all stories for this epic
				storyKeys, err := app.StatusReader.GetEpicStories(epicID)
				if err != nil {
					cmd.SilenceUsage = true
					fmt.Printf("Error reading stories for epic %s: %v\n", epicID, err)
					return NewExitError(1)
				}

				// Execute full lifecycle for each story in order
				for _, storyKey := range storyKeys {
					err := executeWithRetry(ctx, executor, storyKey, autoRetry, 10, func(stepIndex, totalSteps int, workflow string) {
						app.Printer.StepStart(stepIndex, totalSteps, workflow)
					})
					if err != nil {
						cmd.SilenceUsage = true
						if errors.Is(err, router.ErrStoryComplete) {
							fmt.Printf("Story %s is already complete, skipping\n", storyKey)
							continue
						}
						fmt.Printf("Error running lifecycle for story %s: %v\n", storyKey, err)
						return NewExitError(1)
					}
					fmt.Printf("Story %s completed successfully\n", storyKey)
				}

				fmt.Printf("Epic %s completed (%d stories processed)\n\n", epicID, len(storyKeys))
			}

			fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
			fmt.Printf("  All %d epic(s) completed successfully!\n", len(epicIDs))
			fmt.Printf("═══════════════════════════════════════════════════════════════════\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview workflows without executing them")
	cmd.Flags().BoolVar(&autoRetry, "auto-retry", false, "Automatically retry on rate limit errors")

	return cmd
}

func runEpicDryRun(cmd *cobra.Command, app *App, executor *lifecycle.Executor, epicIDs []string) error {
	totalWorkflows := 0
	storiesWithWork := 0
	storiesComplete := 0

	for _, epicID := range epicIDs {
		// Get all stories for this epic
		storyKeys, err := app.StatusReader.GetEpicStories(epicID)
		if err != nil {
			cmd.SilenceUsage = true
			fmt.Printf("Error reading stories for epic %s: %v\n", epicID, err)
			return NewExitError(1)
		}

		fmt.Printf("Epic %s:\n", epicID)

		for _, storyKey := range storyKeys {
			fmt.Printf("  Story %s:\n", storyKey)

			steps, err := executor.GetSteps(storyKey)
			if err != nil {
				if errors.Is(err, router.ErrStoryComplete) {
					fmt.Printf("    (already complete)\n")
					storiesComplete++
					continue
				}
				cmd.SilenceUsage = true
				fmt.Printf("    Error: %v\n", err)
				return NewExitError(1)
			}

			for i, step := range steps {
				modelInfo := ""
				model := app.Config.GetModel(step.Workflow)
				if model != "" {
					modelInfo = fmt.Sprintf(" (%s)", model)
				}
				fmt.Printf("    %d. %s%s → %s\n", i+1, step.Workflow, modelInfo, step.NextStatus)
			}
			totalWorkflows += len(steps)
			storiesWithWork++
		}
		fmt.Println()
	}

	if storiesComplete > 0 {
		fmt.Printf("Total: %d workflows across %d stories (%d already complete)\n", totalWorkflows, storiesWithWork, storiesComplete)
	} else {
		fmt.Printf("Total: %d workflows across %d stories\n", totalWorkflows, storiesWithWork)
	}

	return nil
}
