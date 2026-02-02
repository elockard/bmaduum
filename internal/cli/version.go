package cli

import (
	"github.com/spf13/cobra"
)

// Version information
const (
	Version     = "1.1.0"
	ReleaseDate = "2026-02-02"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Display the version, release date, and other build information for bmad-automate.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("bmad-automate version %s (released %s)\n", Version, ReleaseDate)
		},
	}

	return cmd
}
