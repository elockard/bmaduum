package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newDevStoryCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "dev-story <story-key>",
		Short: "Run dev-story workflow",
		Long:  `Run the dev-story workflow for the specified story key.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			storyKey := args[0]
			ctx := cmd.Context()
			exitCode := app.Runner.RunSingle(ctx, "dev-story", storyKey)
			if exitCode != 0 {
				cmd.SilenceUsage = true
				os.Exit(exitCode)
			}
		},
	}
}
