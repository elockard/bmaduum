package status

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultStatusPath is the default location of the sprint status file.
const DefaultStatusPath = "_bmad-output/implementation-artifacts/sprint-status.yaml"

// Reader reads sprint status from YAML files.
type Reader struct {
	basePath string
}

// NewReader creates a new Reader with the specified base path.
func NewReader(basePath string) *Reader {
	return &Reader{
		basePath: basePath,
	}
}

// Read reads and parses the sprint status file.
func (r *Reader) Read() (*SprintStatus, error) {
	fullPath := filepath.Join(r.basePath, DefaultStatusPath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sprint status: %w", err)
	}

	var status SprintStatus
	if err := yaml.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to read sprint status: %w", err)
	}

	return &status, nil
}

// GetStoryStatus returns the status for a specific story key.
func (r *Reader) GetStoryStatus(storyKey string) (Status, error) {
	sprintStatus, err := r.Read()
	if err != nil {
		return "", err
	}

	status, ok := sprintStatus.DevelopmentStatus[storyKey]
	if !ok {
		return "", fmt.Errorf("story not found: %s", storyKey)
	}

	return status, nil
}
