package vein

import (
	"fmt"
	"os"
	"strings"
)

// Vein defines an injection point for fraglet code
type Vein struct {
	Name       string   `yaml:"name"`                 // Vein name (e.g., "python", "c")
	Container  string   `yaml:"container"`            // Container image (required)
	Extensions []string `yaml:"extensions,omitempty"` // File extensions that map to this vein (e.g., [".py"])
}

// VeinRegistry manages available veins
type VeinRegistry struct {
	veins map[string]*Vein
}

// NewVeinRegistry creates an empty registry
func NewVeinRegistry() *VeinRegistry {
	return &VeinRegistry{
		veins: make(map[string]*Vein),
	}
}

// Add adds a vein to the registry
// Returns an error if a vein with the same name already exists (prevents clobbering)
func (r *VeinRegistry) Add(vein *Vein) error {
	if vein.Name == "" {
		return fmt.Errorf("vein name is required")
	}
	if vein.Container == "" {
		return fmt.Errorf("vein container is required")
	}
	if _, exists := r.veins[vein.Name]; exists {
		return fmt.Errorf("duplicate vein name: %s", vein.Name)
	}
	r.veins[vein.Name] = vein
	return nil
}

// Get retrieves a vein by name
func (r *VeinRegistry) Get(name string) (*Vein, bool) {
	vein, ok := r.veins[name]
	return vein, ok
}

// ContainerImage returns the container image for this vein, applying
// FRAGLET_VEINS_FORCE_TAG if set (e.g. "local" to use locally built images).
func (v *Vein) ContainerImage() string {
	return ApplyForceTag(v.Container)
}

// ApplyForceTag returns container with tag overridden when FRAGLET_VEINS_FORCE_TAG is set.
// Example: 100hellos/python:latest + FRAGLET_VEINS_FORCE_TAG=local → 100hellos/python:local
func ApplyForceTag(container string) string {
	tag := os.Getenv("FRAGLET_VEINS_FORCE_TAG")
	if tag == "" {
		return container
	}
	lastColon := strings.LastIndex(container, ":")
	if lastColon == -1 {
		return container + ":" + tag
	}
	return container[:lastColon] + ":" + tag
}

// List returns all vein names
func (r *VeinRegistry) List() []string {
	names := make([]string, 0, len(r.veins))
	for name := range r.veins {
		names = append(names, name)
	}
	return names
}
