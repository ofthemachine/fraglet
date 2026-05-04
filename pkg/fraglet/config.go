package fraglet

import (
	"fmt"
	"os"

	"github.com/ofthemachine/fraglet/pkg/inject"
	"gopkg.in/yaml.v3"
)

// InjectionConfig is an alias for inject.Config for backward compatibility in entrypoint config
type InjectionConfig = inject.Config

// ModeConfig defines configuration for a specific execution mode
type ModeConfig struct {
	Injection InjectionConfig            `json:"injection" yaml:"injection"`
	Guide     string                     `json:"guide" yaml:"guide"`
	Essence   string                     `json:"essence,omitempty" yaml:"essence,omitempty"`
	Execution *EntrypointExecutionConfig `json:"execution,omitempty" yaml:"execution,omitempty"`
}

// EntrypointExecutionConfig defines code execution settings
type EntrypointExecutionConfig struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// EntrypointConfig describes how to inject, store, and execute fraglets inside a container.
type EntrypointConfig struct {
	FragletTempPath string                `json:"fragletTempPath" yaml:"fragletTempPath"`
	Modes           map[string]ModeConfig `json:"modes" yaml:"modes"`
	// Embed default mode directly in the root for simplicity if no modes are defined
	ModeConfig `yaml:",inline"`
}

// DefaultEntrypointConfig returns default config
func DefaultEntrypointConfig() *EntrypointConfig {
	return &EntrypointConfig{
		FragletTempPath: DefaultFragletTempPath,
		ModeConfig: ModeConfig{
			Injection: InjectionConfig{
				CodePath: DefaultCodePath,
				Match:    DefaultFragletInjectionMatch,
			},
			Guide: DefaultGuidePath,
			Execution: &EntrypointExecutionConfig{
				Path: DefaultCodePath,
			},
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

// LoadEntrypointConfig loads config.
// Priority:
// 1. FRAGLET_CONFIG_PATH (explicit path to config file)
// 2. /fraglet.yml or /fraglet.yaml
// After loading, it selects the mode from FRAGLET_MODE.
func LoadEntrypointConfig() (*EntrypointConfig, error) {
	path := os.Getenv("FRAGLET_CONFIG_PATH")
	if path == "" {
		// Compatibility with old env var name if new one is not set
		path = os.Getenv("FRAGLET_CONFIG")
	}

	if path == "" {
		for _, candidate := range []string{"/fraglet.yml", "/fraglet.yaml"} {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
	}

	var cfg *EntrypointConfig
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		cfg = &EntrypointConfig{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	} else {
		cfg = DefaultEntrypointConfig()
	}

	// Apply defaults to the root config
	defaults := DefaultEntrypointConfig()
	if cfg.FragletTempPath == "" {
		cfg.FragletTempPath = defaults.FragletTempPath
	}
	cfg.ModeConfig = mergeModeConfig(cfg.ModeConfig, defaults.ModeConfig)

	// Resolve mode
	mode := os.Getenv("FRAGLET_MODE")
	if mode != "" {
		if modeCfg, ok := cfg.Modes[mode]; ok {
			// Override root config with mode-specific config
			cfg.ModeConfig = mergeModeConfig(modeCfg, cfg.ModeConfig)
		} else {
			// If mode not found in map, it might be that FRAGLET_CONFIG pointed to a mode-specific file
			// like /fraglet-main.yml in the old system. We've already loaded it into the root.
			// But for strictness, we could warn or error.
			// Given "collapse into single fraglet.yml", we assume the user might still pass the mode name.
		}
	}

	return cfg, nil
}

func mergeModeConfig(target, source ModeConfig) ModeConfig {
	if isEmptyInjection(target.Injection) {
		target.Injection = source.Injection
	} else if target.Injection.CodePath == "" {
		target.Injection.CodePath = source.Injection.CodePath
	}

	if target.Guide == "" {
		target.Guide = source.Guide
	}

	if target.Essence == "" {
		target.Essence = source.Essence
	}

	if target.Execution == nil {
		target.Execution = source.Execution
	} else if target.Execution.Path == "" && source.Execution != nil {
		target.Execution.Path = source.Execution.Path
	}

	return target
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
