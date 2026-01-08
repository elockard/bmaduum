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
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		v: viper.New(),
	}
}

// Load loads configuration from the default locations and environment.
// Priority (highest to lowest):
// 1. Environment variables (BMAD_ prefix)
// 2. Config file specified by BMAD_CONFIG_PATH
// 3. ./config/workflows.yaml
// 4. Default configuration
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Set up Viper
	l.v.SetConfigType("yaml")
	l.v.SetEnvPrefix("BMAD")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()

	// Try to find and read config file
	configPath := os.Getenv("BMAD_CONFIG_PATH")
	if configPath != "" {
		l.v.SetConfigFile(configPath)
	} else {
		// Look for config in current directory
		l.v.SetConfigName("workflows")
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
	if binaryPath := os.Getenv("BMAD_CLAUDE_PATH"); binaryPath != "" {
		cfg.Claude.BinaryPath = binaryPath
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file path.
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
func (c *Config) GetPrompt(workflowName, storyKey string) (string, error) {
	workflow, ok := c.Workflows[workflowName]
	if !ok {
		return "", fmt.Errorf("unknown workflow: %s", workflowName)
	}

	return expandTemplate(workflow.PromptTemplate, PromptData{StoryKey: storyKey})
}

// GetFullCycleSteps returns the list of steps for a full cycle.
func (c *Config) GetFullCycleSteps() []string {
	return c.FullCycle.Steps
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
// Useful for initialization where errors should be fatal.
func MustLoad() *Config {
	loader := NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}
