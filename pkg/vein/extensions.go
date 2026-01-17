package vein

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExtensionMap maps file extensions to vein names
type ExtensionMap struct {
	extToVein     map[string]string   // e.g., ".py" -> "python"
	extConflicts  map[string][]string // e.g., ".m" -> ["mercury", "objective-c", "octave"]
}

// NewExtensionMap creates an extension map from a registry
func NewExtensionMap(registry *VeinRegistry) *ExtensionMap {
	extMap := &ExtensionMap{
		extToVein:    make(map[string]string),
		extConflicts: make(map[string][]string),
	}

	// First pass: collect all extensions and their veins
	extToVeins := make(map[string][]string)
	for _, veinName := range registry.List() {
		vein, _ := registry.Get(veinName)
		for _, ext := range vein.Extensions {
			// Normalize extension (ensure it starts with .)
			normalized := normalizeExtension(ext)
			if normalized != "" {
				extToVeins[normalized] = append(extToVeins[normalized], veinName)
			}
		}
	}

	// Second pass: pick winners for conflicts and build maps
	for ext, veins := range extToVeins {
		if len(veins) > 1 {
			// Conflict detected - sort and pick first alphabetically for determinism
			sortedVeins := make([]string, len(veins))
			copy(sortedVeins, veins)
			sort.Strings(sortedVeins)
			extMap.extToVein[ext] = sortedVeins[0]
			extMap.extConflicts[ext] = sortedVeins
		} else {
			// No conflict
			extMap.extToVein[ext] = veins[0]
		}
	}

	return extMap
}

// VeinForExtension returns the vein name for a given file extension
// Returns error if extension is unknown
// Issues a warning to stderr if the extension has conflicts
func (m *ExtensionMap) VeinForExtension(ext string) (string, error) {
	normalized := normalizeExtension(ext)
	vein, ok := m.extToVein[normalized]
	if !ok {
		return "", fmt.Errorf("unknown extension %s, use --vein to specify", ext)
	}

	// Warn if there's a conflict
	if conflicts, hasConflict := m.extConflicts[normalized]; hasConflict {
		fmt.Fprintf(os.Stderr, "Warning: extension %s is ambiguous (used by: %s). Using '%s'. Specify --vein to override.\n",
			ext, strings.Join(conflicts, ", "), vein)
	}

	return vein, nil
}

// VeinForFile extracts extension from filename and returns the vein name
func (m *ExtensionMap) VeinForFile(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", fmt.Errorf("no extension found in %s, use --vein to specify", filename)
	}
	return m.VeinForExtension(ext)
}

// normalizeExtension ensures extension starts with .
func normalizeExtension(ext string) string {
	ext = strings.ToLower(ext)
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}
