package manifest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadModulesFromFile_Full(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_full.yaml"))

	require.NoError(t, err)
	require.NotNil(t, mm)
	assert.Len(t, mm.Modules, 5)

	assert.Equal(t, "bmm", mm.Modules[0].Name)
	assert.Equal(t, "6.0.0", mm.Modules[0].Version)
	assert.Equal(t, "bmm", mm.Modules[0].Path)

	assert.Equal(t, "sdet", mm.Modules[2].Name)
	assert.Equal(t, "1.0.0", mm.Modules[2].Version)
	assert.Equal(t, "modules/sdet", mm.Modules[2].Path)
}

func TestReadModulesFromFile_Minimal(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_minimal.yaml"))

	require.NoError(t, err)
	require.NotNil(t, mm)
	assert.Len(t, mm.Modules, 1)
	assert.Equal(t, "bmm", mm.Modules[0].Name)
}

func TestReadModulesFromFile_SDET(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_sdet.yaml"))

	require.NoError(t, err)
	assert.Len(t, mm.Modules, 2)
	assert.True(t, mm.HasModule("bmm"))
	assert.True(t, mm.HasModule("sdet"))
	assert.False(t, mm.HasModule("tea"))
}

func TestReadModulesFromFile_TEA(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_tea.yaml"))

	require.NoError(t, err)
	assert.Len(t, mm.Modules, 2)
	assert.True(t, mm.HasModule("bmm"))
	assert.True(t, mm.HasModule("tea"))
	assert.False(t, mm.HasModule("sdet"))
}

func TestReadModulesFromFile_NotFound(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "nonexistent.yaml"))

	assert.Error(t, err)
	assert.Nil(t, mm)
	assert.Contains(t, err.Error(), "failed to read module manifest")
}

func TestReadModulesFromFile_InvalidYAML(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_invalid.yaml"))

	assert.Error(t, err)
	assert.Nil(t, mm)
	assert.Contains(t, err.Error(), "failed to parse module manifest")
}

func TestReadModulesFromFile_EmptyModules(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_empty.yaml"))

	assert.Error(t, err)
	assert.Nil(t, mm)
	assert.Contains(t, err.Error(), "no modules")
}

func TestReadModulesFromFile_NoName(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_no_name.yaml"))

	assert.Error(t, err)
	assert.Nil(t, mm)
	assert.Contains(t, err.Error(), "has no name")
}

func TestReadModulesFromBytes(t *testing.T) {
	data := []byte(`modules:
  - name: bmm
    version: "6.0.0"
  - name: sdet
    version: "1.0.0"
`)
	mm, err := ReadModulesFromBytes(data)

	require.NoError(t, err)
	assert.Len(t, mm.Modules, 2)
	assert.Equal(t, "bmm", mm.Modules[0].Name)
	assert.Equal(t, "sdet", mm.Modules[1].Name)
}

func TestModuleManifest_HasModule(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_full.yaml"))
	require.NoError(t, err)

	assert.True(t, mm.HasModule("bmm"))
	assert.True(t, mm.HasModule("tea"))
	assert.True(t, mm.HasModule("sdet"))
	assert.True(t, mm.HasModule("bmgd"))
	assert.True(t, mm.HasModule("cis"))
	assert.False(t, mm.HasModule("nonexistent"))
	assert.False(t, mm.HasModule(""))
}

func TestModuleManifest_GetModule(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_full.yaml"))
	require.NoError(t, err)

	m := mm.GetModule("sdet")
	require.NotNil(t, m)
	assert.Equal(t, "sdet", m.Name)
	assert.Equal(t, "1.0.0", m.Version)
	assert.Equal(t, "modules/sdet", m.Path)

	m = mm.GetModule("nonexistent")
	assert.Nil(t, m)
}

func TestModuleManifest_Names(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_full.yaml"))
	require.NoError(t, err)

	names := mm.Names()
	assert.Equal(t, []string{"bmgd", "bmm", "cis", "sdet", "tea"}, names)
}

func TestModuleManifest_Names_Minimal(t *testing.T) {
	mm, err := ReadModulesFromFile(filepath.Join("testdata", "modules_minimal.yaml"))
	require.NoError(t, err)

	names := mm.Names()
	assert.Equal(t, []string{"bmm"}, names)
}
