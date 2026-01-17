package vein

import "fmt"

// Vein defines an injection point for fraglet code
type Vein struct {
	Name       string   `yaml:"name"`                 // Vein name (e.g., "python", "the-c-programming-language")
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

// List returns all vein names
func (r *VeinRegistry) List() []string {
	names := make([]string, 0, len(r.veins))
	for name := range r.veins {
		names = append(names, name)
	}
	return names
}
