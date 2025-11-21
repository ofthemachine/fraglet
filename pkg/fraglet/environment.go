package fraglet

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ofthemachine/fraglet/pkg/runner"
)

// FragletEnvironment executes FragletProcs with envelope configs
type FragletEnvironment struct {
	registry *EnvelopeRegistry
}

// NewFragletEnvironment creates an environment with envelope registry
func NewFragletEnvironment(envelopesDir string) (*FragletEnvironment, error) {
	registry, err := NewEnvelopeRegistry(envelopesDir)
	if err != nil {
		return nil, err
	}

	return &FragletEnvironment{registry: registry}, nil
}

// Execute runs a FragletProc using the specified envelope
func (e *FragletEnvironment) Execute(ctx context.Context, envelopeName string, proc FragletProc) (*runner.RunResult, error) {
	// Resolve envelope
	envelope, ok := e.registry.GetEnvelope(envelopeName)
	if !ok {
		return nil, fmt.Errorf("envelope not found: %s", envelopeName)
	}

	// Handle two cases: fraglet-entrypoint vs bare container
	if envelope.IsFragletEntrypointEnabled() {
		return e.executeWithFragletEntrypoint(ctx, envelope, proc)
	} else {
		return e.executeWithBareContainer(ctx, envelope, proc)
	}
}

// GetRegistry returns the envelope registry (for tools that need direct access)
func (e *FragletEnvironment) GetRegistry() *EnvelopeRegistry {
	return e.registry
}

// executeWithFragletEntrypoint handles fraglet-entrypoint enabled containers
func (e *FragletEnvironment) executeWithFragletEntrypoint(ctx context.Context, envelope *FragletEnvelope, proc FragletProc) (*runner.RunResult, error) {
	// Write code to temp file
	tmpFile, cleanup, err := writeTempFile(proc.Code())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Create runner
	r := runner.NewRunner(envelope.Container, "")

	// Execute with volume mount at fragletPath
	spec := runner.RunSpec{
		Container: envelope.Container,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: envelope.FragletPath,
				ReadOnly:      true,
			},
		},
	}

	result, err := r.Run(ctx, spec)
	return &result, err
}

// executeWithBareContainer handles bare containers (with template injection)
func (e *FragletEnvironment) executeWithBareContainer(ctx context.Context, envelope *FragletEnvelope, proc FragletProc) (*runner.RunResult, error) {
	// Inject fraglet code into template
	rendered, err := InjectString(envelope.FragletTemplate, proc.Code(), envelope.Injection)
	if err != nil {
		return nil, fmt.Errorf("template injection failed: %w", err)
	}

	// Write rendered code to temp file
	tmpFile, cleanup, err := writeTempFile(rendered)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Make executable if needed
	if envelope.Execution != nil && envelope.Execution.MakeExecutable {
		if err := os.Chmod(tmpFile, 0755); err != nil {
			return nil, fmt.Errorf("failed to make file executable: %w", err)
		}
	}

	// Create runner
	r := runner.NewRunner(envelope.Container, "")

	// Execute with volume mount and explicit command
	spec := runner.RunSpec{
		Container: envelope.Container,
		Command:   envelope.Execution.Path,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: envelope.FragletPath,
				ReadOnly:      false, // May need to make executable
			},
		},
	}

	result, err := r.Run(ctx, spec)
	return &result, err
}

func writeTempFile(content string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "fraglet-*")
	if err != nil {
		return "", nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0644)

	absPath, _ := filepath.Abs(tmpFile.Name())
	cleanup := func() { os.Remove(absPath) }

	return absPath, cleanup, nil
}
