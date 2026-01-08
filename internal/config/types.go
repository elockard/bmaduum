// Package config provides configuration loading and management for bmad-automate.
package config

// Config represents the root configuration structure.
type Config struct {
	Workflows map[string]WorkflowConfig `mapstructure:"workflows"`
	FullCycle FullCycleConfig           `mapstructure:"full_cycle"`
	Claude    ClaudeConfig              `mapstructure:"claude"`
	Output    OutputConfig              `mapstructure:"output"`
}

// WorkflowConfig represents a single workflow configuration.
type WorkflowConfig struct {
	PromptTemplate string `mapstructure:"prompt_template"`
}

// FullCycleConfig defines the steps for a full development cycle.
type FullCycleConfig struct {
	Steps []string `mapstructure:"steps"`
}

// ClaudeConfig contains Claude CLI configuration.
type ClaudeConfig struct {
	OutputFormat string `mapstructure:"output_format"`
	BinaryPath   string `mapstructure:"binary_path"`
}

// OutputConfig contains output formatting configuration.
type OutputConfig struct {
	TruncateLines  int `mapstructure:"truncate_lines"`
	TruncateLength int `mapstructure:"truncate_length"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Workflows: map[string]WorkflowConfig{
			"create-story": {
				PromptTemplate: "/bmad:bmm:workflows:create-story - Create story: {{.StoryKey}}. Do not ask questions.",
			},
			"dev-story": {
				PromptTemplate: "/bmad:bmm:workflows:dev-story - Work on story: {{.StoryKey}}. Complete all tasks. Run tests after each implementation. Do not ask clarifying questions - use best judgment based on existing patterns.",
			},
			"code-review": {
				PromptTemplate: "/bmad:bmm:workflows:code-review - Review story: {{.StoryKey}}. When presenting fix options, always choose to auto-fix all issues immediately. Do not wait for user input.",
			},
			"git-commit": {
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
		},
	}
}

// PromptData contains data for template expansion.
type PromptData struct {
	StoryKey string
}
