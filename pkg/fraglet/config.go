package fraglet

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EntrypointConfig describes how to inject, store, and execute fraglets inside a container.
type EntrypointConfig struct {
	FragletTempPath string                     `json:"fragletTempPath" yaml:"fragletTempPath"`
	Injection       InjectionConfig            `json:"injection" yaml:"injection"`
	AgentHelp       string                     `json:"agentHelp" yaml:"agentHelp"` // Path to agent-help markdown file (mode is implicit: container)
	HowTo           string                     `json:"howTo" yaml:"howTo"`         // Path to how-to markdown file
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
		AgentHelp: DefaultAgentHelpPath,
		HowTo:     DefaultHowToPath,
		Execution: &EntrypointExecutionConfig{
			Path:           DefaultCodePath,
			MakeExecutable: &makeExec,
		},
	}
}

const (
	DefaultEntrypointConfigPath  = "/fraglet-entrypoint.yaml"
	DefaultCodePath              = "/code/hello-world.sh"
	DefaultFragletTempPath       = "/FRAGLET"
	DefaultFragletInjectionMatch = "FRAGLET"
	DefaultAgentHelpPath         = "/agent-help.md"
	DefaultHowToPath             = "/how-to.md"
)

// LoadEntrypointConfig loads config from FRAGLET_CONFIG envvar path, or looks for
// fraglet-entrypoint.yaml as a sibling to the binary, or uses default path
func LoadEntrypointConfig() (*EntrypointConfig, error) {
	path := os.Getenv("FRAGLET_CONFIG")
	if path == "" {
		// Try to find config as sibling to the binary
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			siblingPath := filepath.Join(execDir, "fraglet-entrypoint.yaml")
			if _, err := os.Stat(siblingPath); err == nil {
				path = siblingPath
			}
		}

		// If still not found, try default path
		if path == "" {
			path = DefaultEntrypointConfigPath
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// No config file, return defaults
				return DefaultEntrypointConfig(), nil
			}
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
	if cfg.AgentHelp == "" {
		cfg.AgentHelp = defaults.AgentHelp
	}
	if cfg.HowTo == "" {
		cfg.HowTo = defaults.HowTo
	}
	if cfg.Execution == nil {
		cfg.Execution = defaults.Execution
	} else if cfg.Execution.MakeExecutable == nil {
		cfg.Execution.MakeExecutable = defaults.Execution.MakeExecutable
	}

	return &cfg, nil
}

func isEmptyInjection(inj InjectionConfig) bool {
	return inj.Match == "" && inj.MatchStart == "" && inj.MatchEnd == ""
}
