package entrypoint

import (
	"fmt"
	"os"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// Manager handles fraglet injection workflow
type Manager struct {
	cfg *fraglet.EntrypointConfig
}

// NewManager creates a new fraglet manager
func NewManager(cfg *fraglet.EntrypointConfig) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

// Process handles the complete fraglet injection process. In argv mode
// (Execution.Argv) there is no scaffold file to inject into — the mounted
// fraglet body is read and exec'd directly by the executor — so injection
// is skipped entirely; the file is left in place at FragletTempPath for the
// executor to read.
func (m *Manager) Process() error {
	if m.cfg.Execution != nil && m.cfg.Execution.Argv {
		return nil
	}

	fragletPath := m.cfg.FragletTempPath
	if fragletPath == "" {
		return nil
	}
	if _, err := os.Stat(fragletPath); os.IsNotExist(err) {
		return nil
	}

	injector := NewInjector()
	if err := injector.Inject(fragletPath, m.cfg.Injection); err != nil {
		return fmt.Errorf("error injecting fraglet: %w", err)
	}

	return nil
}
