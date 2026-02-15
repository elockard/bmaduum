package manifest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFromFile_Valid(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Len(t, m.Entries, 5)

	// Check first entry
	assert.Equal(t, "3", m.Entries[0].Phase)
	assert.Equal(t, "create-story", m.Entries[0].Workflow)
	assert.Equal(t, "SM", m.Entries[0].Agent)
	assert.Equal(t, "/create-story", m.Entries[0].Command)
	assert.Equal(t, "backlog", m.Entries[0].TriggerStatus)
	assert.Equal(t, "ready-for-dev", m.Entries[0].NextStatus)

	// Check git-commit (no trigger status)
	assert.Equal(t, "git-commit", m.Entries[4].Workflow)
	assert.Equal(t, "", m.Entries[4].TriggerStatus)
	assert.Equal(t, "done", m.Entries[4].NextStatus)
}

func TestReadFromFile_Minimal(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "minimal.csv"))

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Len(t, m.Entries, 2)

	// Minimal CSV only has required columns
	assert.Equal(t, "dev-story", m.Entries[0].Workflow)
	assert.Equal(t, "ready-for-dev", m.Entries[0].TriggerStatus)
	assert.Equal(t, "review", m.Entries[0].NextStatus)
	assert.Equal(t, "", m.Entries[0].Phase)
	assert.Equal(t, "", m.Entries[0].Agent)
	assert.Equal(t, "", m.Entries[0].Command)
}

func TestReadFromFile_NotFound(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "nonexistent.csv"))

	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to open manifest")
}

func TestReadFromFile_MissingColumn(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "missing_column.csv"))

	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "missing required column")
}

func TestReadFromFile_EmptyWorkflow(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "empty_workflow.csv"))

	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "workflow name is required")
}

func TestReadFromFile_HeaderOnly(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "header_only.csv"))

	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "no workflow entries")
}

func TestReadFromString(t *testing.T) {
	csv := `phase,workflow,agent,command,trigger_status,next_status
3,create-story,SM,/create-story,backlog,ready-for-dev
3,dev-story,Dev,/dev-story,ready-for-dev,review
`
	m, err := ReadFromString(csv)

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Len(t, m.Entries, 2)
	assert.Equal(t, "create-story", m.Entries[0].Workflow)
	assert.Equal(t, "dev-story", m.Entries[1].Workflow)
}

func TestReadFromString_Empty(t *testing.T) {
	m, err := ReadFromString("")

	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to read manifest header")
}

func TestManifest_Workflows(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)

	workflows := m.Workflows()

	// dev-story appears twice but should only be listed once
	assert.Equal(t, []string{"create-story", "dev-story", "code-review", "git-commit"}, workflows)
}

func TestManifest_GetWorkflowEntry(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)

	entry := m.GetWorkflowEntry("dev-story")
	require.NotNil(t, entry)
	assert.Equal(t, "Dev", entry.Agent)
	assert.Equal(t, "/dev-story", entry.Command)

	// Returns first match (ready-for-dev trigger)
	assert.Equal(t, "ready-for-dev", entry.TriggerStatus)
}

func TestManifest_GetWorkflowEntry_NotFound(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)

	entry := m.GetWorkflowEntry("nonexistent")
	assert.Nil(t, entry)
}

func TestManifest_HasWorkflow(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)

	assert.True(t, m.HasWorkflow("create-story"))
	assert.True(t, m.HasWorkflow("dev-story"))
	assert.True(t, m.HasWorkflow("code-review"))
	assert.True(t, m.HasWorkflow("git-commit"))
	assert.False(t, m.HasWorkflow("nonexistent"))
}

func TestManifest_GetEntriesForStatus(t *testing.T) {
	m, err := ReadFromFile(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)

	entries := m.GetEntriesForStatus("backlog")
	assert.Len(t, entries, 1)
	assert.Equal(t, "create-story", entries[0].Workflow)

	entries = m.GetEntriesForStatus("ready-for-dev")
	assert.Len(t, entries, 1)
	assert.Equal(t, "dev-story", entries[0].Workflow)

	entries = m.GetEntriesForStatus("in-progress")
	assert.Len(t, entries, 1)
	assert.Equal(t, "dev-story", entries[0].Workflow)

	entries = m.GetEntriesForStatus("review")
	assert.Len(t, entries, 1)
	assert.Equal(t, "code-review", entries[0].Workflow)

	// No entries for "done" or unknown statuses
	entries = m.GetEntriesForStatus("done")
	assert.Len(t, entries, 0)

	entries = m.GetEntriesForStatus("unknown")
	assert.Len(t, entries, 0)
}

func TestReadFromString_TrimsWhitespace(t *testing.T) {
	csv := `phase, workflow, agent, command, trigger_status, next_status
3, create-story, SM, /create-story, backlog, ready-for-dev
`
	m, err := ReadFromString(csv)

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "create-story", m.Entries[0].Workflow)
	assert.Equal(t, "SM", m.Entries[0].Agent)
	assert.Equal(t, "backlog", m.Entries[0].TriggerStatus)
}
