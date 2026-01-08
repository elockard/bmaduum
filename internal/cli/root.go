// Package cli provides the command-line interface for bmad-automate.
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"bmad-automate/internal/claude"
	"bmad-automate/internal/config"
	"bmad-automate/internal/output"
	"bmad-automate/internal/workflow"
)

// App holds the application dependencies.
type App struct {
	Config   *config.Config
	Executor claude.Executor
	Printer  output.Printer
	Runner   *workflow.Runner
	Queue    *workflow.QueueRunner
}

// NewApp creates a new application with all dependencies wired up.
func NewApp(cfg *config.Config) *App {
	printer := output.NewPrinter()

	executor := claude.NewExecutor(claude.ExecutorConfig{
		BinaryPath:   cfg.Claude.BinaryPath,
		OutputFormat: cfg.Claude.OutputFormat,
		StderrHandler: func(line string) {
			// Print stderr to stderr
			os.Stderr.WriteString("[stderr] " + line + "\n")
		},
	})

	runner := workflow.NewRunner(executor, printer, cfg)
	queue := workflow.NewQueueRunner(runner)

	return &App{
		Config:   cfg,
		Executor: executor,
		Printer:  printer,
		Runner:   runner,
		Queue:    queue,
	}
}

// NewRootCommand creates the root cobra command.
func NewRootCommand(app *App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bmad-automate",
		Short: "BMAD Automation CLI",
		Long: `BMAD Automation CLI - Automate development workflows with Claude.

This tool orchestrates Claude to run development workflows including
story creation, development, code review, and git operations.`,
	}

	// Add subcommands
	rootCmd.AddCommand(
		newCreateStoryCommand(app),
		newDevStoryCommand(app),
		newCodeReviewCommand(app),
		newGitCommitCommand(app),
		newRunCommand(app),
		newQueueCommand(app),
		newRawCommand(app),
	)

	return rootCmd
}

// Execute runs the CLI application.
func Execute() {
	// Load configuration
	cfg, err := config.NewLoader().Load()
	if err != nil {
		os.Stderr.WriteString("Error loading config: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Create app and root command
	app := NewApp(cfg)
	rootCmd := NewRootCommand(app)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
