package fraglet

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FragletEnvelope defines execution context for fraglet code
type FragletEnvelope struct {
	Name            string           `yaml:"name"`
	Language        string           `yaml:"language"`
	Container       string           `yaml:"container"`
	FragletPath     string           `yaml:"fragletPath"`               // Where to inject/mount code
	FragletConfig   string           `yaml:"fragletConfig,omitempty"`   // Optional: path to fraglet config file (for fraglet-entrypoint containers)
	FragletTemplate string           `yaml:"fragletTemplate,omitempty"` // Optional: template with markers (bare containers)
	Injection       *InjectionConfig `yaml:"injection,omitempty"`       // Optional: injection pattern config (bare containers)
	Guide           string           `yaml:"guide,omitempty"`           // Optional: path to guide file (bare containers only; fraglet-entrypoint provides via "guide" command)
	Execution       *ExecutionConfig `yaml:"execution,omitempty"`       // Optional: only for bare containers
}

// InjectionConfig defines how to inject fraglet code into a target file
type InjectionConfig struct {
	CodePath   string `yaml:"codePath"`              // Full path to target file where injection occurs
	Match      string `yaml:"match,omitempty"`       // Simple string match (replaced with fraglet code)
	MatchStart string `yaml:"match_start,omitempty"` // Start marker (region between match_start and match_end is replaced)
	MatchEnd   string `yaml:"match_end,omitempty"`   // End marker (used with match_start)
}

// ExecutionConfig defines how code should be executed (bare containers only)
type ExecutionConfig struct {
	Path           string `yaml:"path"` // Path to execute (e.g., /tmp/fraglet_exec.py)
	MakeExecutable bool   `yaml:"makeExecutable,omitempty"`
}

// IsFragletEntrypointEnabled returns true if this envelope uses fraglet-entrypoint
// (i.e., no fragletTemplate and no execution path)
func (e *FragletEnvelope) IsFragletEntrypointEnabled() bool {
	return e.FragletTemplate == "" && (e.Execution == nil || e.Execution.Path == "")
}

// EnvelopeRegistry manages available envelopes
type EnvelopeRegistry struct {
	envelopes map[string]*FragletEnvelope
}

// NewEnvelopeRegistry creates a registry and loads envelopes from dir
func NewEnvelopeRegistry(envelopesDir string) (*EnvelopeRegistry, error) {
	r := &EnvelopeRegistry{envelopes: make(map[string]*FragletEnvelope)}

	files, err := filepath.Glob(filepath.Join(envelopesDir, "*.yml"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		env, err := loadEnvelope(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", file, err)
		}
		r.envelopes[env.Name] = env
	}

	return r, nil
}

// NewEnvelopeRegistryFromEmbedded creates a registry and loads envelopes from embedded filesystem
func NewEnvelopeRegistryFromEmbedded() (*EnvelopeRegistry, error) {
	r := &EnvelopeRegistry{envelopes: make(map[string]*FragletEnvelope)}

	envelopesFS := getEmbeddedEnvelopesFS()
	err := fs.WalkDir(envelopesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yml" {
			return nil
		}

		data, err := fs.ReadFile(envelopesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded envelope %s: %w", path, err)
		}

		var env FragletEnvelope
		if err := yaml.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("failed to parse embedded envelope %s: %w", path, err)
		}

		r.envelopes[env.Name] = &env
		return nil
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

// GetEnvelope retrieves envelope by name
func (r *EnvelopeRegistry) GetEnvelope(name string) (*FragletEnvelope, bool) {
	env, ok := r.envelopes[name]
	return env, ok
}

// ListEnvelopes returns all envelope names
func (r *EnvelopeRegistry) ListEnvelopes() []string {
	names := make([]string, 0, len(r.envelopes))
	for name := range r.envelopes {
		names = append(names, name)
	}
	return names
}

func loadEnvelope(path string) (*FragletEnvelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var env FragletEnvelope
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, err
	}

	return &env, nil
}
