package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
)

// Loader handles configuration loading from files and environment.
//
// Loader uses Viper to load configuration from YAML files and environment
// variables, merging them with default values. The loader supports the
// BMADUUM_ environment variable prefix for all configuration options.
type Loader struct {
	// v is the Viper instance used for configuration loading.
	v *viper.Viper
}

// NewLoader creates a new configuration loader.
//
// Returns a Loader ready to load configuration from files and environment.
// Call [Loader.Load] to perform the actual configuration loading.
func NewLoader() *Loader {
	return &Loader{
		v: viper.New(),
	}
}

// Load loads configuration from the default locations and environment.
//
// Configuration is loaded and merged with the following priority (highest first):
//  1. Environment variables with BMADUUM_ prefix (e.g., BMADUUM_CLAUDE_BINARY_PATH)
//  2. Config file specified by BMADUUM_CONFIG_PATH environment variable
//  3. User config directory: ~/.config/bmaduum/workflows.yaml (Linux),
//     ~/Library/Application Support/bmaduum/workflows.yaml (macOS),
//     %APPDATA%\bmaduum\workflows.yaml (Windows)
//  4. ./config/workflows.yaml in the current directory (legacy fallback)
//  5. ./workflows.yaml in the current directory (legacy fallback)
//  6. [DefaultConfig] built-in defaults
//
// Environment variable names use underscores for nested keys. For example,
// claude.binary_path becomes BMADUUM_CLAUDE_BINARY_PATH.
//
// Returns an error if a config file exists but cannot be parsed. Missing
// config files are not an error; the loader falls back to defaults.
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Set up Viper
	l.v.SetConfigType("yaml")
	l.v.SetEnvPrefix("BMADUUM")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()

	// Try to find and read config file
	configPath := os.Getenv("BMADUUM_CONFIG_PATH")
	if configPath != "" {
		// User explicitly set a config path
		l.v.SetConfigFile(configPath)
	} else {
		// Search for config in multiple locations (in priority order)
		l.v.SetConfigName("workflows")

		// 1. User config directory (platform-standard location)
		userConfigDir, err := os.UserConfigDir()
		if err == nil {
			appConfigDir := filepath.Join(userConfigDir, "bmaduum")
			l.v.AddConfigPath(appConfigDir)
		}

		// 2. Legacy local directories (for backward compatibility)
		l.v.AddConfigPath("./config")
		l.v.AddConfigPath(".")
	}

	// Read config file if it exists
	if err := l.v.ReadInConfig(); err != nil {
		// Config file not found is okay, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Some other error occurred
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into config struct
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override Claude binary path from env if set
	if binaryPath := os.Getenv("BMADUUM_CLAUDE_PATH"); binaryPath != "" {
		cfg.Claude.BinaryPath = binaryPath
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file path.
//
// Unlike [Loader.Load], this method loads from an explicit file path without
// searching default locations or checking environment variables. The file
// extension determines the expected format (yaml, json, etc.).
//
// Returns an error if the file cannot be read or parsed.
func (l *Loader) LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	l.v.SetConfigFile(path)
	l.v.SetConfigType(filepath.Ext(path)[1:]) // Remove the dot

	if err := l.v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return cfg, nil
}

// GetPrompt returns the expanded prompt for a workflow and story key.
//
// The workflowName must match a key in the Workflows map. The storyKey is
// substituted into the workflow's prompt template using Go's text/template.
//
// When [Config.UseSlashCommands] is true (default), the workflow's SlashCommand
// template is used (e.g., "/dev-story {{.StoryKey}}"). When false, the legacy
// PromptTemplate is used instead. If the selected template is empty, the other
// template is used as a fallback.
//
// Returns an error if the workflow is not found or if template expansion fails.
func (c *Config) GetPrompt(workflowName, storyKey string) (string, error) {
	workflow, ok := c.Workflows[workflowName]
	if !ok {
		return "", fmt.Errorf("unknown workflow: %s", workflowName)
	}

	tmpl := workflow.SlashCommand
	if !c.UseSlashCommands {
		tmpl = workflow.PromptTemplate
	}

	// Fallback: if the selected template is empty, use the other one
	if tmpl == "" {
		if c.UseSlashCommands {
			tmpl = workflow.PromptTemplate
		} else {
			tmpl = workflow.SlashCommand
		}
	}

	if tmpl == "" {
		return "", fmt.Errorf("workflow %s has no prompt template or slash command configured", workflowName)
	}

	return expandTemplate(tmpl, PromptData{StoryKey: storyKey})
}

// GetFullCycleSteps returns the list of workflow steps for a full lifecycle.
//
// This returns the configured FullCycle.Steps slice, which defines the
// sequence of workflows to execute for run, queue, and epic commands.
func (c *Config) GetFullCycleSteps() []string {
	return c.FullCycle.Steps
}

// GetModel returns the model configured for a workflow, or empty string if not set.
//
// When empty, the Claude CLI will use its default model.
func (c *Config) GetModel(workflowName string) string {
	workflow, ok := c.Workflows[workflowName]
	if !ok {
		return ""
	}
	return workflow.Model
}

// expandTemplate expands a Go template string with the given data.
func expandTemplate(tmpl string, data PromptData) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
}

// MustLoad loads configuration and panics on error.
//
// This is a convenience function for initialization code where configuration
// errors should be fatal. It creates a new [Loader] and calls [Loader.Load],
// panicking if an error occurs.
//
// Use this in main() or package initialization where there is no reasonable
// way to handle configuration errors.
func MustLoad() *Config {
	loader := NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// ConfigDir returns the platform-specific user configuration directory for bmaduum.
//
// The returned path follows platform conventions:
//   - Linux: ~/.config/bmaduum
//   - macOS: ~/Library/Application Support/bmaduum
//   - Windows: %APPDATA%\bmaduum
//
// Returns an error if the user config directory cannot be determined.
func ConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user config directory: %w", err)
	}
	return filepath.Join(userConfigDir, "bmaduum"), nil
}

// DefaultConfigPath returns the full path to the default config file.
//
// This is the path where workflows.yaml will be looked for by default
// (after checking BMAD_CONFIG_PATH environment variable).
//
// Returns an error if the user config directory cannot be determined.
func DefaultConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "workflows.yaml"), nil
}

// EnsureConfigDir ensures the user configuration directory exists.
//
// If the directory does not exist, it will be created with appropriate
// permissions (0755). If the directory already exists, this is a no-op.
//
// Returns an error if the directory cannot be created.
func EnsureConfigDir() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(configDir, 0755)
}
