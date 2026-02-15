// Package config provides configuration loading and management for bmaduum.
//
// Configuration is loaded using Viper, supporting YAML config files and environment
// variable overrides. The package provides sensible defaults that work out of the
// box, with the ability to customize workflows, output formatting, and Claude CLI
// settings.
//
// Key types:
//   - [Config] is the root configuration container with all settings
//   - [Loader] handles Viper-based configuration loading
//   - [WorkflowConfig] defines a single workflow's prompt template
//   - [ClaudeConfig] contains Claude CLI binary settings
//
// Configuration priority (highest to lowest):
//  1. Environment variables (BMADUUM_ prefix)
//  2. Config file specified by BMADUUM_CONFIG_PATH
//  3. User config directory (platform-standard):
//     - Linux: ~/.config/bmaduum/workflows.yaml
//     - macOS: ~/Library/Application Support/bmaduum/workflows.yaml
//     - Windows: %APPDATA%\bmaduum\workflows.yaml
//  4. ./config/workflows.yaml (legacy fallback)
//  5. ./workflows.yaml (legacy fallback)
//  6. [DefaultConfig] defaults
package config

// Config represents the root configuration structure.
//
// This is the main configuration container loaded by [Loader] and used throughout
// the application. Use [DefaultConfig] to get sensible defaults.
type Config struct {
	// Workflows maps workflow names to their configurations.
	// Keys are workflow names (e.g., "create-story", "dev-story").
	Workflows map[string]WorkflowConfig `mapstructure:"workflows"`

	// UseSlashCommands controls whether workflows invoke BMAD v6 slash commands
	// or use legacy prompt templates. When true (default), GetPrompt returns
	// the SlashCommand template. When false, it returns the PromptTemplate.
	// Set to false for pre-v6 BMAD projects that use /bmad-bmm-* commands.
	UseSlashCommands bool `mapstructure:"use_slash_commands"`

	// FullCycle defines the steps for full lifecycle execution.
	// Used by run, queue, and epic commands.
	FullCycle FullCycleConfig `mapstructure:"full_cycle"`

	// Claude contains Claude CLI binary configuration.
	Claude ClaudeConfig `mapstructure:"claude"`

	// Output contains terminal output formatting configuration.
	Output OutputConfig `mapstructure:"output"`
}

// WorkflowConfig represents a single workflow configuration.
//
// Each workflow has two prompt modes: a SlashCommand for BMAD v6 projects
// and a legacy PromptTemplate for older projects. The active mode is
// controlled by [Config.UseSlashCommands].
type WorkflowConfig struct {
	// SlashCommand is the BMAD v6 slash command template.
	// Used when Config.UseSlashCommands is true (default).
	// Example: "/dev-story {{.StoryKey}}"
	SlashCommand string `mapstructure:"slash_command"`

	// PromptTemplate is the legacy Go template string for the workflow prompt.
	// Used when Config.UseSlashCommands is false.
	// Example: "/bmad-bmm-dev-story - Work on story: {{.StoryKey}}"
	PromptTemplate string `mapstructure:"prompt_template"`

	// Model is the Claude model to use for this workflow.
	// If empty, the default model is used.
	// Examples: "opus", "sonnet", "haiku", "claude-sonnet-4-5-20250929"
	Model string `mapstructure:"model"`
}

// FullCycleConfig defines the steps for a full development cycle.
//
// This configuration is used by the run, queue, and epic commands
// to determine the sequence of workflows to execute.
type FullCycleConfig struct {
	// Steps is the ordered list of workflow names to execute.
	// Default: ["create-story", "dev-story", "code-review", "git-commit"]
	Steps []string `mapstructure:"steps"`
}

// ClaudeConfig contains Claude CLI configuration.
//
// These settings control how the Claude CLI binary is invoked.
type ClaudeConfig struct {
	// OutputFormat is the output format passed to Claude CLI.
	// Should be "stream-json" for structured event parsing.
	OutputFormat string `mapstructure:"output_format"`

	// BinaryPath is the path to the Claude CLI binary.
	// Default: "claude" (assumes Claude is in PATH).
	// Can be overridden with BMADUUM_CLAUDE_PATH environment variable.
	BinaryPath string `mapstructure:"binary_path"`
}

// OutputConfig contains terminal output formatting configuration.
//
// These settings control how Claude's output is formatted in the terminal.
type OutputConfig struct {
	// TruncateLines is the maximum number of lines to display per event.
	// Additional lines are hidden with a "... (N more lines)" indicator.
	// Default: 20
	TruncateLines int `mapstructure:"truncate_lines"`

	// TruncateLength is the maximum length of each output line.
	// Longer lines are truncated with "..." suffix.
	// Default: 60
	TruncateLength int `mapstructure:"truncate_length"`

	// Markdown contains markdown rendering configuration.
	Markdown MarkdownConfig `mapstructure:"markdown"`
}

// MarkdownConfig contains configuration for markdown rendering in terminal output.
//
// When enabled, Claude's text output is rendered with proper formatting:
// bold, italic, headers, code blocks with syntax highlighting, lists, etc.
type MarkdownConfig struct {
	// Enabled controls whether markdown rendering is active.
	// Default: true
	Enabled bool `mapstructure:"enabled"`

	// Style is the glamour theme to use: "dark", "light", "dracula", "tokyo-night".
	// Avoid "auto" as it can cause detection delays on some terminals.
	// Default: "dark"
	Style string `mapstructure:"style"`

	// WordWrap is the column width for text wrapping.
	// Default: 100
	WordWrap int `mapstructure:"word_wrap"`

	// Emoji enables emoji shortcode rendering (e.g., :smile: -> 😄).
	// Default: true
	Emoji bool `mapstructure:"emoji"`
}

// DefaultConfig returns a new [Config] with sensible defaults.
//
// The defaults include standard workflow prompts for create-story, dev-story,
// code-review, and git-commit workflows, as well as Claude CLI and output
// formatting settings. These defaults work out of the box without any
// configuration file.
func DefaultConfig() *Config {
	return &Config{
		UseSlashCommands: true,
		Workflows: map[string]WorkflowConfig{
			"create-story": {
				SlashCommand:   "/create-story {{.StoryKey}}",
				PromptTemplate: "/bmad-bmm-create-story - Create story: {{.StoryKey}}. Do not ask questions.",
			},
			"dev-story": {
				SlashCommand:   "/dev-story {{.StoryKey}}",
				PromptTemplate: "/bmad-bmm-dev-story - Work on story: {{.StoryKey}}. Complete all tasks. Run tests after each implementation. Do not ask clarifying questions - use best judgment based on existing patterns.",
			},
			"code-review": {
				SlashCommand:   "/code-review {{.StoryKey}}",
				PromptTemplate: "/bmad-bmm-code-review - Review story: {{.StoryKey}}. When presenting fix options, always choose to auto-fix all issues immediately. Do not wait for user input.",
			},
			"git-commit": {
				SlashCommand:   "/git-commit {{.StoryKey}}",
				PromptTemplate: "Commit all changes for story {{.StoryKey}} with a descriptive commit message following conventional commits format. Then push to the current branch. Do not ask questions.",
			},
		},
		FullCycle: FullCycleConfig{
			Steps: []string{"create-story", "dev-story", "code-review", "git-commit"},
		},
		Claude: ClaudeConfig{
			OutputFormat: "stream-json",
			BinaryPath:   "claude",
		},
		Output: OutputConfig{
			TruncateLines:  20,
			TruncateLength: 60,
			Markdown: MarkdownConfig{
				Enabled:  true,
				Style:    "dark",
				WordWrap: 100,
				Emoji:    true,
			},
		},
	}
}

// PromptData contains data for workflow template expansion.
//
// This struct is passed to Go's text/template when expanding workflow prompts.
// Fields are accessible in templates using {{.FieldName}} syntax.
type PromptData struct {
	// StoryKey is the identifier of the story being processed.
	// Access in templates with {{.StoryKey}}.
	StoryKey string
}
