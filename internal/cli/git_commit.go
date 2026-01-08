package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newGitCommitCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "git-commit <story-key>",
		Short: "Commit and push changes for a story",
		Long:  `Commit all changes for the specified story with a descriptive commit message and push to the current branch.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			storyKey := args[0]
			ctx := cmd.Context()
			exitCode := app.Runner.RunSingle(ctx, "git-commit", storyKey)
			if exitCode != 0 {
				cmd.SilenceUsage = true
				os.Exit(exitCode)
			}
		},
	}
}
