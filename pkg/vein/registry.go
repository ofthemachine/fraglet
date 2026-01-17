package vein

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// VeinsDirEnvVar is the environment variable that, when set, makes the system
	// use filesystem veins instead of embedded ones. This is useful for development.
	VeinsDirEnvVar = "FRAGLET_VEINS_DIR"
)

// VeinsConfig represents the top-level structure of veins.yml
type VeinsConfig struct {
	Veins []*Vein `yaml:"veins"`
}

// LoadFromFile loads veins from a single veins.yml file
func LoadFromFile(path string) (*VeinRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read veins file: %w", err)
	}

	var config VeinsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse veins file: %w", err)
	}

	registry := NewVeinRegistry()
	for _, vein := range config.Veins {
		if err := registry.Add(vein); err != nil {
			return nil, fmt.Errorf("invalid vein %s: %w", vein.Name, err)
		}
	}

	return registry, nil
}

// LoadFromDir loads veins from all YAML files in a directory
// Reads all .yml and .yaml files and combines their veins into a single registry
func LoadFromDir(dir string) (*VeinRegistry, error) {
	registry := NewVeinRegistry()

	// Find all YAML files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var foundFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".yml" || filepath.Ext(name) == ".yaml" {
			foundFiles = append(foundFiles, filepath.Join(dir, name))
		}
	}

	if len(foundFiles) == 0 {
		return nil, fmt.Errorf("no YAML files found in %s", dir)
	}

	// Load veins from each file
	for _, filePath := range foundFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		var config VeinsConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		// Add all veins from this file to the registry
		for _, vein := range config.Veins {
			if err := registry.Add(vein); err != nil {
				return nil, fmt.Errorf("failed to add vein from %s: %w", filePath, err)
			}
		}
	}

	return registry, nil
}

// LoadAuto loads veins, checking FRAGLET_VEINS_DIR first, then falling back to embedded
func LoadAuto(loadEmbedded func() (*VeinRegistry, error)) (*VeinRegistry, error) {
	veinsDir := os.Getenv(VeinsDirEnvVar)
	if veinsDir != "" {
		// Use filesystem veins (development mode)
		return LoadFromDir(veinsDir)
	}
	// Use embedded veins (production mode)
	return loadEmbedded()
}
