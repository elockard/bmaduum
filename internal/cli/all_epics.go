package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"bmad-automate/internal/lifecycle"
	"bmad-automate/internal/router"
)

func newAllEpicsCommand(app *App) *cobra.Command {
	var dryRun bool
	var autoRetry bool

	cmd := &cobra.Command{
		Use:   "all-epics",
		Short: "Run lifecycle for all active epics",
		Long: `Run the complete lifecycle for all active epics in sequence.

This command automatically discovers all epics and processes them in order.
Each epic's stories are executed through their full lifecycle from current status to done.

Active epics are those with stories that are not marked as "done", "deferred", or "optional".
Epics are processed in numerical order (epic-1, epic-2, etc.).

Use --dry-run to preview all workflows without executing them.
Use --auto-retry to automatically retry on rate limit errors.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Get all epics
			epics, err := app.StatusReader.GetAllEpics()
			if err != nil {
				cmd.SilenceUsage = true
				fmt.Printf("Error reading epics: %v\n", err)
				return NewExitError(1)
			}

			if len(epics) == 0 {
				fmt.Println("No epics found")
				return nil
			}

			fmt.Printf("Found %d epic(s): %v\n\n", len(epics), epics)

			// Process each epic
			for epicIndex, epicID := range epics {
				fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
				fmt.Printf("  Epic %d of %d: epic-%s\n", epicIndex+1, len(epics), epicID)
				fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")

				// Get stories for this epic
				stories, err := app.StatusReader.GetEpicStories(epicID)
				if err != nil {
					cmd.SilenceUsage = true
					fmt.Printf("Error reading stories for epic %s: %v\n", epicID, err)
					return NewExitError(1)
				}

				// Process each story in the epic
				for storyIndex, storyKey := range stories {
					fmt.Printf("─── Story %d of %d: %s\n", storyIndex+1, len(stories), storyKey)

					// Create lifecycle executor
					executor := lifecycle.NewExecutor(app.Runner, app.StatusReader, app.StatusWriter)

					// Handle dry-run mode
					if dryRun {
						steps, err := executor.GetSteps(storyKey)
						if err != nil {
							if errors.Is(err, router.ErrStoryComplete) {
								fmt.Printf("  Story is already complete, skipping\n\n")
								continue
							}
							fmt.Printf("  Error: %v\n\n", err)
							continue
						}

						fmt.Printf("  Would execute %d workflow(s):\n", len(steps))
						for i, step := range steps {
							modelInfo := ""
							model := app.Config.GetModel(step.Workflow)
							if model != "" {
								modelInfo = fmt.Sprintf(" (%s)", model)
							}
							fmt.Printf("    %d. %s%s → %s\n", i+1, step.Workflow, modelInfo, step.NextStatus)
						}
						fmt.Println()
						continue
					}

					// Execute with optional retry
					err := executeWithRetry(ctx, executor, storyKey, autoRetry, 10, func(stepIndex, totalSteps int, workflow string) {
						app.Printer.StepStart(stepIndex, totalSteps, workflow)
					})

					if err != nil {
						cmd.SilenceUsage = true
						if errors.Is(err, router.ErrStoryComplete) {
							fmt.Printf("Story %s is already complete, skipping\n\n", storyKey)
							continue
						}
						fmt.Printf("Error processing story %s: %v\n", storyKey, err)
						return NewExitError(1)
					}

					fmt.Printf("Story %s completed successfully\n\n", storyKey)
				}

				fmt.Printf("Epic %s completed (%d stories processed)\n\n", epicID, len(stories))
			}

			fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
			fmt.Printf("  All %d epic(s) completed successfully!\n", len(epics))
			fmt.Printf("═══════════════════════════════════════════════════════════════════\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview workflows without executing them")
	cmd.Flags().BoolVar(&autoRetry, "auto-retry", false, "Automatically retry on rate limit errors")

	return cmd
}
