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

	// Check slash commands enabled by default
	assert.True(t, cfg.UseSlashCommands)

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

func TestLoader_LoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
workflows:
  custom-workflow:
    prompt_template: "Custom: {{.StoryKey}}"
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
	assert.Equal(t, "/custom/path/claude", cfg.Claude.BinaryPath)
	assert.Equal(t, 50, cfg.Output.TruncateLines)
}

func TestLoader_Load_WithEnvOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("BMADUUM_CLAUDE_PATH", "/env/claude")
	defer os.Unsetenv("BMADUUM_CLAUDE_PATH")

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

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	assert.NotNil(t, loader)
	assert.NotNil(t, loader.v)
}

func TestLoader_LoadFromFile_NonExistent(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("/nonexistent/path/config.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading config file")
}

func TestLoader_LoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidContent := `
workflows:
  - this is not valid yaml for this structure
    missing: colon here
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.LoadFromFile(configPath)

	// Should error on unmarshal due to wrong structure
	assert.Error(t, err)
}

func TestLoader_Load_DefaultsWithNoConfigFile(t *testing.T) {
	// Ensure no config file exists in current dir
	// Load() should fall back to defaults
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Clear any env vars that might interfere
	os.Unsetenv("BMADUUM_CONFIG_PATH")
	os.Unsetenv("BMADUUM_CLAUDE_PATH")

	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	// Should have default values
	assert.Equal(t, "claude", cfg.Claude.BinaryPath)
	assert.Equal(t, "stream-json", cfg.Claude.OutputFormat)
}

func TestLoader_Load_WithConfigPathEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yaml")

	configContent := `
claude:
  binary_path: /from/env/path/claude
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	os.Setenv("BMADUUM_CONFIG_PATH", configPath)
	defer os.Unsetenv("BMADUUM_CONFIG_PATH")

	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, "/from/env/path/claude", cfg.Claude.BinaryPath)
}

func TestLoader_Load_EnvOverridesTakePrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config file sets one path
	configContent := `
claude:
  binary_path: /from/file/claude
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	os.Setenv("BMADUUM_CONFIG_PATH", configPath)
	os.Setenv("BMADUUM_CLAUDE_PATH", "/from/env/override/claude")
	defer os.Unsetenv("BMADUUM_CONFIG_PATH")
	defer os.Unsetenv("BMADUUM_CLAUDE_PATH")

	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	// Env var should take precedence
	assert.Equal(t, "/from/env/override/claude", cfg.Claude.BinaryPath)
}

func TestMustLoad_Success(t *testing.T) {
	// MustLoad should not panic when loading defaults
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	os.Unsetenv("BMADUUM_CONFIG_PATH")
	os.Unsetenv("BMADUUM_CLAUDE_PATH")

	// Should not panic
	cfg := MustLoad()
	assert.NotNil(t, cfg)
}

func TestConfig_GetPrompt_AllWorkflows(t *testing.T) {
	cfg := DefaultConfig()

	workflows := []string{"create-story", "dev-story", "code-review", "git-commit"}

	for _, wf := range workflows {
		t.Run(wf, func(t *testing.T) {
			prompt, err := cfg.GetPrompt(wf, "test-key")
			assert.NoError(t, err)
			assert.NotEmpty(t, prompt)
			assert.Contains(t, prompt, "test-key")
		})
	}
}

func TestLoader_LoadFromFile_DifferentExtension(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Write valid JSON config
	jsonContent := `{
		"claude": {
			"binary_path": "/json/path/claude"
		}
	}`
	err := os.WriteFile(configPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.Equal(t, "/json/path/claude", cfg.Claude.BinaryPath)
}

func TestDefaultConfig_WorkflowTemplates(t *testing.T) {
	cfg := DefaultConfig()

	// Verify each workflow has both slash command and legacy template
	for name, workflow := range cfg.Workflows {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, workflow.PromptTemplate, "workflow %s should have a legacy template", name)
			assert.NotEmpty(t, workflow.SlashCommand, "workflow %s should have a slash command", name)
		})
	}
}

func TestPromptData_StoryKey(t *testing.T) {
	data := PromptData{StoryKey: "ABC-123"}
	assert.Equal(t, "ABC-123", data.StoryKey)
}

func TestConfigDir(t *testing.T) {
	configDir, err := ConfigDir()
	require.NoError(t, err)
	assert.NotEmpty(t, configDir)
	assert.Contains(t, configDir, "bmaduum")
}

func TestDefaultConfigPath(t *testing.T) {
	configPath, err := DefaultConfigPath()
	require.NoError(t, err)
	assert.NotEmpty(t, configPath)
	assert.Contains(t, configPath, "bmaduum")
	assert.Contains(t, configPath, "workflows.yaml")
}

func TestEnsureConfigDir(t *testing.T) {
	// We can't easily mock os.UserConfigDir, but we can verify
	// that EnsureConfigDir creates a directory when given a path
	// Let's test the logic by checking that a real call doesn't error
	err := EnsureConfigDir()
	// This should either succeed or fail due to permissions,
	// but not due to logic errors
	// We can't make strong assertions since we're using the real OS config dir
	if err != nil {
		// If it fails, it should be a permission error or similar
		// Not a "not implemented" error
		assert.NotContains(t, err.Error(), "not implemented")
	}
}

func TestGetPrompt_SlashCommandMode(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.UseSlashCommands)

	prompt, err := cfg.GetPrompt("dev-story", "6-1-implement-auth")
	require.NoError(t, err)
	assert.Equal(t, "/dev-story 6-1-implement-auth", prompt)
}

func TestGetPrompt_LegacyMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.UseSlashCommands = false

	prompt, err := cfg.GetPrompt("dev-story", "6-1-implement-auth")
	require.NoError(t, err)
	assert.Contains(t, prompt, "/bmad-bmm-dev-story")
	assert.Contains(t, prompt, "6-1-implement-auth")
}

func TestGetPrompt_SlashCommandMode_AllWorkflows(t *testing.T) {
	cfg := DefaultConfig()

	expected := map[string]string{
		"create-story": "/create-story test-key",
		"dev-story":    "/dev-story test-key",
		"code-review":  "/code-review test-key",
		"git-commit":   "/git-commit test-key",
	}

	for wf, want := range expected {
		t.Run(wf, func(t *testing.T) {
			prompt, err := cfg.GetPrompt(wf, "test-key")
			assert.NoError(t, err)
			assert.Equal(t, want, prompt)
		})
	}
}

func TestGetPrompt_LegacyMode_AllWorkflows(t *testing.T) {
	cfg := DefaultConfig()
	cfg.UseSlashCommands = false

	for _, wf := range []string{"create-story", "dev-story", "code-review", "git-commit"} {
		t.Run(wf, func(t *testing.T) {
			prompt, err := cfg.GetPrompt(wf, "test-key")
			assert.NoError(t, err)
			assert.Contains(t, prompt, "test-key")
			// Legacy prompts should NOT start with the v6 slash command
			assert.NotEqual(t, "/"+wf+" test-key", prompt)
		})
	}
}

func TestGetPrompt_FallbackToLegacy(t *testing.T) {
	// When UseSlashCommands is true but SlashCommand is empty, fall back to PromptTemplate
	cfg := &Config{
		UseSlashCommands: true,
		Workflows: map[string]WorkflowConfig{
			"custom": {
				PromptTemplate: "Legacy: {{.StoryKey}}",
				// SlashCommand intentionally empty
			},
		},
	}

	prompt, err := cfg.GetPrompt("custom", "key-1")
	require.NoError(t, err)
	assert.Equal(t, "Legacy: key-1", prompt)
}

func TestGetPrompt_FallbackToSlashCommand(t *testing.T) {
	// When UseSlashCommands is false but PromptTemplate is empty, fall back to SlashCommand
	cfg := &Config{
		UseSlashCommands: false,
		Workflows: map[string]WorkflowConfig{
			"custom": {
				SlashCommand: "/custom {{.StoryKey}}",
				// PromptTemplate intentionally empty
			},
		},
	}

	prompt, err := cfg.GetPrompt("custom", "key-1")
	require.NoError(t, err)
	assert.Equal(t, "/custom key-1", prompt)
}

func TestGetPrompt_BothEmpty(t *testing.T) {
	cfg := &Config{
		UseSlashCommands: true,
		Workflows: map[string]WorkflowConfig{
			"empty": {},
		},
	}

	_, err := cfg.GetPrompt("empty", "key-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no prompt template or slash command configured")
}

func TestLoader_LoadFromFile_WithSlashCommands(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
use_slash_commands: true
workflows:
  dev-story:
    slash_command: "/dev-story {{.StoryKey}}"
    prompt_template: "Legacy dev: {{.StoryKey}}"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.True(t, cfg.UseSlashCommands)
	assert.Equal(t, "/dev-story {{.StoryKey}}", cfg.Workflows["dev-story"].SlashCommand)
	assert.Equal(t, "Legacy dev: {{.StoryKey}}", cfg.Workflows["dev-story"].PromptTemplate)
}

func TestLoader_LoadFromFile_LegacyMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
use_slash_commands: false
workflows:
  dev-story:
    slash_command: "/dev-story {{.StoryKey}}"
    prompt_template: "Legacy dev: {{.StoryKey}}"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.False(t, cfg.UseSlashCommands)

	prompt, err := cfg.GetPrompt("dev-story", "test-key")
	require.NoError(t, err)
	assert.Equal(t, "Legacy dev: test-key", prompt)
}
