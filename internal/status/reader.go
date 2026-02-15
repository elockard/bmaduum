package status

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// V6StatusPath is the BMAD v6 canonical location of sprint-status.yaml
// relative to the project root.
const V6StatusPath = "_bmad-output/implementation-artifacts/sprint-status.yaml"

// LegacyStatusPath is the pre-v6 location of sprint-status.yaml at the
// project root.
const LegacyStatusPath = "sprint-status.yaml"

// DefaultStatusPath is an alias for V6StatusPath for backward compatibility.
const DefaultStatusPath = V6StatusPath

// StatusPaths lists the paths to search (in priority order) when auto-discovering
// the sprint-status.yaml file.
var StatusPaths = []string{
	V6StatusPath,
	LegacyStatusPath,
}

// ResolvePath discovers the sprint-status.yaml file location.
//
// Resolution order:
//  1. BMADUUM_SPRINT_STATUS_PATH environment variable (used as-is if set)
//  2. Explicit statusPath parameter (if non-empty)
//  3. Auto-discovery: tries v6 path, then legacy path under basePath
//  4. Falls back to v6 path (will error on read if file doesn't exist)
//
// The basePath is the project root directory. Pass empty string for cwd.
// The statusPath is an explicit override (e.g., from config). Pass empty
// string for auto-discovery.
func ResolvePath(basePath, statusPath string) string {
	// 1. Environment variable takes highest priority
	if envPath := os.Getenv("BMADUUM_SPRINT_STATUS_PATH"); envPath != "" {
		return envPath
	}

	// 2. Explicit path from config
	if statusPath != "" {
		return statusPath
	}

	// 3. Auto-discover by checking each path
	for _, p := range StatusPaths {
		fullPath := filepath.Join(basePath, p)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// 4. Default to v6 path
	return filepath.Join(basePath, V6StatusPath)
}

// Reader reads sprint status from YAML files.
//
// The statusPath field holds the resolved full path to the sprint-status.yaml file.
// Use [NewReader] for auto-discovery or [NewReaderWithPath] for an explicit path.
type Reader struct {
	statusPath string
}

// NewReader creates a new [Reader] that auto-discovers the status file.
//
// The basePath is the project root directory. Pass an empty string to use
// the current working directory. The reader searches for sprint-status.yaml
// at the v6 path first, then falls back to the legacy root-level path.
// The BMADUUM_SPRINT_STATUS_PATH environment variable overrides all discovery.
func NewReader(basePath string) *Reader {
	return &Reader{
		statusPath: ResolvePath(basePath, ""),
	}
}

// NewReaderWithPath creates a new [Reader] that uses the specified status file path.
//
// The statusPath can be an absolute path or a path relative to the working directory.
// The BMADUUM_SPRINT_STATUS_PATH environment variable still takes priority if set.
func NewReaderWithPath(basePath, statusPath string) *Reader {
	return &Reader{
		statusPath: ResolvePath(basePath, statusPath),
	}
}

// Read reads and parses the complete sprint status file.
//
// It returns the full [SprintStatus] structure containing all story statuses.
// Returns an error if the file cannot be read or parsed.
func (r *Reader) Read() (*SprintStatus, error) {
	fullPath := r.statusPath

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

// GetStoryStatus returns the [Status] for a specific story key.
//
// It reads the status file and looks up the given story. Returns an error
// if the file cannot be read or if the story key is not found in the file.
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

// GetEpicStories returns all story keys belonging to an epic, sorted by story number.
//
// Story keys are matched using the pattern {epicID}-{N}-*, where N is a numeric
// story number. Results are sorted numerically by story number (1, 2, 10 not 1, 10, 2).
//
// Returns an error if the file cannot be read or if no stories are found for the epic.
func (r *Reader) GetEpicStories(epicID string) ([]string, error) {
	sprintStatus, err := r.Read()
	if err != nil {
		return nil, err
	}

	// Collect all keys matching the epic ID pattern
	type storyWithNum struct {
		key string
		num int
	}
	var stories []storyWithNum

	prefix := epicID + "-"
	for key := range sprintStatus.DevelopmentStatus {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Extract the story number (second segment)
		// Format: {epicID}-{storyNum}-{rest}
		remainder := strings.TrimPrefix(key, prefix)
		parts := strings.SplitN(remainder, "-", 2)
		if len(parts) < 1 {
			continue
		}

		num, err := strconv.Atoi(parts[0])
		if err != nil {
			// Not a numeric story number, skip
			continue
		}

		stories = append(stories, storyWithNum{key: key, num: num})
	}

	if len(stories) == 0 {
		return nil, fmt.Errorf("no stories found for epic: %s", epicID)
	}

	// Sort by story number
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].num < stories[j].num
	})

	// Extract just the keys
	result := make([]string, len(stories))
	for i, s := range stories {
		result[i] = s.key
	}

	return result, nil
}

// GetAllEpics returns all epic IDs with active status, sorted numerically.
//
// An epic is considered "active" if its status is not "done", "deferred", or "optional".
// Epic IDs are extracted from story keys (format: {epicID}-{storyNum}-{description}).
// Results are sorted numerically by epic ID.
func (r *Reader) GetAllEpics() ([]string, error) {
	sprintStatus, err := r.Read()
	if err != nil {
		return nil, err
	}

	// Collect unique epic IDs with their numeric value for sorting
	type epicInfo struct {
		id  string
		num int
	}

	epicMap := make(map[string]int)

	for key := range sprintStatus.DevelopmentStatus {
		// Extract epic ID from story key (format: {epicID}-{storyNum}-{description})
		parts := strings.SplitN(key, "-", 2)
		if len(parts) < 1 {
			continue
		}

		epicID := parts[0]

		// Check if this epic is active
		// For now, we return all epics - the caller can filter by status if needed
		epicMap[epicID] = 0
	}

	if len(epicMap) == 0 {
		return []string{}, nil
	}

	// Convert to slice and parse numeric values for sorting
	epics := make([]epicInfo, 0, len(epicMap))
	for id := range epicMap {
		num, _ := strconv.Atoi(id)
		epics = append(epics, epicInfo{id: id, num: num})
	}

	// Sort by numeric value
	sort.Slice(epics, func(i, j int) bool {
		return epics[i].num < epics[j].num
	})

	// Extract just the IDs
	result := make([]string, len(epics))
	for i, e := range epics {
		result[i] = e.id
	}

	return result, nil
}
