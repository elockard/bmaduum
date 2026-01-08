package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check workflows exist
	assert.Contains(t, cfg.Workflows, "create-story")
	assert.Contains(t, cfg.Workflows, "dev-story")
	assert.Contains(t, cfg.Workflows, "code-review")
	assert.Contains(t, cfg.Workflows, "git-commit")

	// Check full cycle steps
	assert.Equal(t, []string{"create-story", "dev-story", "code-review", "git-commit"}, cfg.FullCycle.Steps)

	// Check defaults
	assert.Equal(t, "stream-json", cfg.Claude.OutputFormat)
	assert.Equal(t, "claude", cfg.Claude.BinaryPath)
	assert.Equal(t, 20, cfg.Output.TruncateLines)
	assert.Equal(t, 60, cfg.Output.TruncateLength)
}

func TestConfig_GetPrompt(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name         string
		workflowName string
		storyKey     string
		wantContains string
		wantErr      bool
	}{
		{
			name:         "create-story",
			workflowName: "create-story",
			storyKey:     "test-123",
			wantContains: "test-123",
			wantErr:      false,
		},
		{
			name:         "dev-story",
			workflowName: "dev-story",
			storyKey:     "feature-456",
			wantContains: "feature-456",
			wantErr:      false,
		},
		{
			name:         "unknown workflow",
			workflowName: "unknown",
			storyKey:     "test",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := cfg.GetPrompt(tt.workflowName, tt.storyKey)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, prompt, tt.wantContains)
			}
		})
	}
}

func TestConfig_GetFullCycleSteps(t *testing.T) {
	cfg := DefaultConfig()
	steps := cfg.GetFullCycleSteps()

	assert.Equal(t, []string{"create-story", "dev-story", "code-review", "git-commit"}, steps)
}

func TestLoader_LoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
workflows:
  custom-workflow:
    prompt_template: "Custom: {{.StoryKey}}"
full_cycle:
  steps:
    - custom-workflow
claude:
  binary_path: /custom/path/claude
output:
  truncate_lines: 50
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.Contains(t, cfg.Workflows, "custom-workflow")
	assert.Equal(t, []string{"custom-workflow"}, cfg.FullCycle.Steps)
	assert.Equal(t, "/custom/path/claude", cfg.Claude.BinaryPath)
	assert.Equal(t, 50, cfg.Output.TruncateLines)
}

func TestLoader_Load_WithEnvOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("BMAD_CLAUDE_PATH", "/env/claude")
	defer os.Unsetenv("BMAD_CLAUDE_PATH")

	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, "/env/claude", cfg.Claude.BinaryPath)
}

func TestExpandTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     PromptData
		want     string
		wantErr  bool
	}{
		{
			name:     "simple substitution",
			template: "Story: {{.StoryKey}}",
			data:     PromptData{StoryKey: "test-123"},
			want:     "Story: test-123",
			wantErr:  false,
		},
		{
			name:     "multiple substitutions",
			template: "{{.StoryKey}} - {{.StoryKey}}",
			data:     PromptData{StoryKey: "abc"},
			want:     "abc - abc",
			wantErr:  false,
		},
		{
			name:     "no substitution",
			template: "Static text",
			data:     PromptData{StoryKey: "ignored"},
			want:     "Static text",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "{{.Invalid",
			data:     PromptData{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandTemplate(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}
