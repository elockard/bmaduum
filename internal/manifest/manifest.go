// Package manifest reads BMAD v6 workflow manifest files.
//
// The workflow manifest (typically at _bmad/_cfg/workflow-manifest.csv) catalogs
// all installed workflows with their phase, agent, slash command, and routing
// information. This enables dynamic workflow discovery instead of hardcoding
// the status-to-workflow chain.
//
// CSV format:
//
//	phase,workflow,agent,command,trigger_status,next_status
//	3,create-story,SM,/create-story,backlog,ready-for-dev
//	3,dev-story,Dev,/dev-story,ready-for-dev,review
//	3,dev-story,Dev,/dev-story,in-progress,review
//	3,code-review,QA,/code-review,review,done
//	3,git-commit,,/git-commit,,done
//
// Rows are ordered by lifecycle execution sequence. A workflow may appear
// multiple times with different trigger_status values (e.g., dev-story
// triggers on both ready-for-dev and in-progress).
package manifest

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// WorkflowEntry represents a single row in the workflow manifest CSV.
//
// Each entry describes a workflow with its BMAD metadata and routing information.
// A workflow may have multiple entries with different TriggerStatus values.
type WorkflowEntry struct {
	// Phase is the BMAD phase number (e.g., "3" for Implementation).
	Phase string

	// Workflow is the workflow name, matching keys in config/workflows.yaml.
	Workflow string

	// Agent is the BMAD agent responsible for this workflow (e.g., "SM", "Dev", "QA").
	Agent string

	// Command is the slash command to invoke (e.g., "/create-story").
	Command string

	// TriggerStatus is the story status that triggers this workflow.
	// Empty for workflows that are only part of the lifecycle chain (e.g., git-commit).
	TriggerStatus string

	// NextStatus is the status to set after successful workflow completion.
	NextStatus string
}

// Manifest holds all workflow entries parsed from a manifest CSV file.
type Manifest struct {
	// Entries are the workflow entries in lifecycle execution order.
	Entries []WorkflowEntry
}

// ReadFromFile reads and parses a workflow manifest CSV file.
func ReadFromFile(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest: %w", err)
	}
	defer f.Close()

	return readFromReader(f)
}

// ReadFromString parses a workflow manifest from a CSV string.
// This is useful for testing and for embedding manifest data.
func ReadFromString(data string) (*Manifest, error) {
	return readFromReader(strings.NewReader(data))
}

func readFromReader(r io.Reader) (*Manifest, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest header: %w", err)
	}

	colIndex := buildColumnIndex(header)
	if err := validateColumns(colIndex); err != nil {
		return nil, err
	}

	var entries []WorkflowEntry
	lineNum := 1 // header was line 1
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest line %d: %w", lineNum, err)
		}

		entry := WorkflowEntry{
			Phase:         getField(record, colIndex, "phase"),
			Workflow:      getField(record, colIndex, "workflow"),
			Agent:         getField(record, colIndex, "agent"),
			Command:       getField(record, colIndex, "command"),
			TriggerStatus: getField(record, colIndex, "trigger_status"),
			NextStatus:    getField(record, colIndex, "next_status"),
		}

		if entry.Workflow == "" {
			return nil, fmt.Errorf("manifest line %d: workflow name is required", lineNum)
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("manifest contains no workflow entries")
	}

	return &Manifest{Entries: entries}, nil
}

// requiredColumns are the columns that must be present in the manifest CSV.
var requiredColumns = []string{"workflow", "trigger_status", "next_status"}

func buildColumnIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, col := range header {
		index[strings.TrimSpace(strings.ToLower(col))] = i
	}
	return index
}

func validateColumns(colIndex map[string]int) error {
	for _, col := range requiredColumns {
		if _, ok := colIndex[col]; !ok {
			return fmt.Errorf("manifest missing required column: %s", col)
		}
	}
	return nil
}

func getField(record []string, colIndex map[string]int, column string) string {
	idx, ok := colIndex[column]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

// Workflows returns the unique workflow names in lifecycle order.
// The order is determined by first appearance in the manifest.
func (m *Manifest) Workflows() []string {
	seen := make(map[string]bool)
	var workflows []string
	for _, e := range m.Entries {
		if !seen[e.Workflow] {
			seen[e.Workflow] = true
			workflows = append(workflows, e.Workflow)
		}
	}
	return workflows
}

// GetWorkflowEntry returns the first entry matching the given workflow name.
// Returns nil if not found.
func (m *Manifest) GetWorkflowEntry(name string) *WorkflowEntry {
	for _, e := range m.Entries {
		if e.Workflow == name {
			return &e
		}
	}
	return nil
}

// HasWorkflow returns true if the manifest contains the given workflow.
func (m *Manifest) HasWorkflow(name string) bool {
	return m.GetWorkflowEntry(name) != nil
}

// GetEntriesForStatus returns all entries that have the given trigger status.
func (m *Manifest) GetEntriesForStatus(triggerStatus string) []WorkflowEntry {
	var entries []WorkflowEntry
	for _, e := range m.Entries {
		if e.TriggerStatus == triggerStatus {
			entries = append(entries, e)
		}
	}
	return entries
}
