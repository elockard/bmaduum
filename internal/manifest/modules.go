package manifest

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// Module represents an installed BMAD module.
type Module struct {
	// Name is the module identifier (e.g., "bmm", "tea", "sdet", "bmgd", "cis").
	Name string `yaml:"name"`

	// Version is the module version string (e.g., "6.0.0").
	Version string `yaml:"version"`

	// Path is the module's relative path within the _bmad directory.
	Path string `yaml:"path"`
}

// moduleManifestFile represents the raw YAML structure of _bmad/_config/manifest.yaml.
type moduleManifestFile struct {
	Modules []Module `yaml:"modules"`
}

// ModuleManifest holds discovered BMAD modules.
type ModuleManifest struct {
	// Modules is the list of installed modules.
	Modules []Module
}

// ReadModulesFromFile reads and parses a BMAD module manifest YAML file.
//
// The expected file location is _bmad/_config/manifest.yaml relative to the
// project root. The YAML format is:
//
//	modules:
//	  - name: bmm
//	    version: "6.0.0"
//	    path: bmm
//	  - name: sdet
//	    version: "1.0.0"
//	    path: modules/sdet
func ReadModulesFromFile(path string) (*ModuleManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read module manifest: %w", err)
	}

	return ReadModulesFromBytes(data)
}

// ReadModulesFromBytes parses a BMAD module manifest from YAML bytes.
func ReadModulesFromBytes(data []byte) (*ModuleManifest, error) {
	var raw moduleManifestFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse module manifest: %w", err)
	}

	if len(raw.Modules) == 0 {
		return nil, fmt.Errorf("module manifest contains no modules")
	}

	// Validate each module has a name
	for i, m := range raw.Modules {
		if m.Name == "" {
			return nil, fmt.Errorf("module at index %d has no name", i)
		}
	}

	return &ModuleManifest{Modules: raw.Modules}, nil
}

// HasModule returns true if a module with the given name is installed.
// The name comparison is case-sensitive.
func (mm *ModuleManifest) HasModule(name string) bool {
	for _, m := range mm.Modules {
		if m.Name == name {
			return true
		}
	}
	return false
}

// GetModule returns the module with the given name, or nil if not found.
func (mm *ModuleManifest) GetModule(name string) *Module {
	for _, m := range mm.Modules {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

// Names returns all installed module names in sorted order.
func (mm *ModuleManifest) Names() []string {
	names := make([]string, len(mm.Modules))
	for i, m := range mm.Modules {
		names[i] = m.Name
	}
	sort.Strings(names)
	return names
}
