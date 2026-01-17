package fraglet

import (
	"fmt"
	"os"

	"github.com/ofthemachine/fraglet/pkg/inject"
	"gopkg.in/yaml.v3"
)

// InjectionConfig is an alias for inject.Config for backward compatibility in entrypoint config
type InjectionConfig = inject.Config

// EntrypointConfig describes how to inject, store, and execute fraglets inside a container.
type EntrypointConfig struct {
	FragletTempPath string                     `json:"fragletTempPath" yaml:"fragletTempPath"`
	Injection       InjectionConfig            `json:"injection" yaml:"injection"`
	Guide           string                     `json:"guide" yaml:"guide"` // Path to guide markdown file (mode is implicit: container)
	Execution       *EntrypointExecutionConfig `json:"execution,omitempty" yaml:"execution,omitempty"`
}

// EntrypointExecutionConfig defines code execution settings
type EntrypointExecutionConfig struct {
	Path           string `json:"path,omitempty" yaml:"path,omitempty"`
	MakeExecutable *bool  `json:"makeExecutable,omitempty" yaml:"makeExecutable,omitempty"`
}

// ShouldMakeExecutable returns whether files should be made executable, defaulting to true
func (e *EntrypointExecutionConfig) ShouldMakeExecutable() bool {
	if e == nil || e.MakeExecutable == nil {
		return true
	}
	return *e.MakeExecutable
}

// DefaultEntrypointConfig returns default config
func DefaultEntrypointConfig() *EntrypointConfig {
	makeExec := true
	return &EntrypointConfig{
		FragletTempPath: DefaultFragletTempPath,
		Injection: InjectionConfig{
			CodePath: DefaultCodePath,
			Match:    DefaultFragletInjectionMatch,
		},
		Guide: DefaultGuidePath,
		Execution: &EntrypointExecutionConfig{
			Path:           DefaultCodePath,
			MakeExecutable: &makeExec,
		},
	}
}

const (
	DefaultEntrypointConfigPath  = "/fraglet.yaml"
	DefaultCodePath              = "/code/hello-world.sh"
	DefaultFragletTempPath       = "/FRAGLET"
	DefaultFragletInjectionMatch = "FRAGLET"
	DefaultGuidePath             = "/guide.md"
)

// LoadEntrypointConfig loads config using mode convention:
// - If FRAGLET_CONFIG is set, use that path
// - Otherwise, try /fraglet.yml or /fraglet.yaml (default mode)
// - Falls back to defaults if no config found
func LoadEntrypointConfig() (*EntrypointConfig, error) {
	path := os.Getenv("FRAGLET_CONFIG")
	if path == "" {
		// Default mode convention: try /fraglet.yml then /fraglet.yaml
		for _, candidate := range []string{"/fraglet.yml", "/fraglet.yaml"} {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}

		// If still not found, return defaults
		if path == "" {
			return DefaultEntrypointConfig(), nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg EntrypointConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	defaults := DefaultEntrypointConfig()
	if cfg.FragletTempPath == "" {
		cfg.FragletTempPath = defaults.FragletTempPath
	}
	if isEmptyInjection(cfg.Injection) {
		cfg.Injection = defaults.Injection
	} else if cfg.Injection.CodePath == "" {
		cfg.Injection.CodePath = defaults.Injection.CodePath
	}
	if cfg.Guide == "" {
		cfg.Guide = defaults.Guide
	}
	if cfg.Execution == nil {
		cfg.Execution = defaults.Execution
	} else if cfg.Execution.MakeExecutable == nil {
		cfg.Execution.MakeExecutable = defaults.Execution.MakeExecutable
	}

	return &cfg, nil
}

// isLineInjection returns true when a single-line marker should be replaced.
func isLineInjection(inj InjectionConfig) bool {
	return inj.Match != ""
}

// isRangeInjection returns true when a region between start/end markers is replaced.
func isRangeInjection(inj InjectionConfig) bool {
	return inj.MatchStart != "" && inj.MatchEnd != ""
}

// isFileInjection returns true when the entire file should be replaced (codePath only).
func isFileInjection(inj InjectionConfig) bool {
	return inj.CodePath != "" && !isLineInjection(inj) && !isRangeInjection(inj)
}

func isEmptyInjection(inj InjectionConfig) bool {
	return !isFileInjection(inj) && !isLineInjection(inj) && !isRangeInjection(inj)
}
