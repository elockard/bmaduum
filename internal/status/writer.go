package status

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Writer writes sprint status to YAML files.
type Writer struct {
	basePath string
}

// NewWriter creates a new Writer with the specified base path.
func NewWriter(basePath string) *Writer {
	return &Writer{
		basePath: basePath,
	}
}

// UpdateStatus updates the status for a specific story key in sprint-status.yaml.
func (w *Writer) UpdateStatus(storyKey string, newStatus Status) error {
	// Validate the new status
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	fullPath := filepath.Join(w.basePath, DefaultStatusPath)

	// Read existing file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read sprint status: %w", err)
	}

	// Parse YAML
	var status SprintStatus
	if err := yaml.Unmarshal(data, &status); err != nil {
		return fmt.Errorf("failed to parse sprint status: %w", err)
	}

	// Check if story exists
	if _, ok := status.DevelopmentStatus[storyKey]; !ok {
		return fmt.Errorf("story not found: %s", storyKey)
	}

	// Update status
	status.DevelopmentStatus[storyKey] = newStatus

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(&status)
	if err != nil {
		return fmt.Errorf("failed to marshal sprint status: %w", err)
	}

	// Write back to file atomically (write to temp, then rename)
	tmpPath := fullPath + ".tmp"
	if err := os.WriteFile(tmpPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write sprint status: %w", err)
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write sprint status: %w", err)
	}

	return nil
}
