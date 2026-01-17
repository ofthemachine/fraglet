package inject

import (
	"fmt"
	"os"
	"strings"
)

// Config defines how to inject fraglet code into a target file
type Config struct {
	CodePath   string `yaml:"codePath"`              // Full path to target file where injection occurs
	Match      string `yaml:"match,omitempty"`       // Simple string match (replaced with fraglet code)
	MatchStart string `yaml:"match_start,omitempty"` // Start marker (region between match_start and match_end is replaced)
	MatchEnd   string `yaml:"match_end,omitempty"`   // End marker (used with match_start)
}

// InjectString injects fraglet code into a template string using injection config.
// The first line containing the match string (as a substring) is used as the injection point.
func InjectString(template string, fragletCode string, config *Config) (string, error) {
	if config == nil {
		return "", fmt.Errorf("injection config is required")
	}

	if config.Match != "" {
		return injectWithMatchRegion(template, fragletCode, config.Match, config.Match)
	}

	if config.MatchStart != "" && config.MatchEnd != "" {
		return injectWithMatchRegion(template, fragletCode, config.MatchStart, config.MatchEnd)
	}

	return "", fmt.Errorf("invalid injection config: must provide match or match_start/match_end")
}

// InjectFile injects fraglet code into a target file using injection config.
// This is a file-level wrapper around InjectString that handles IO operations.
// The target file path comes from config.CodePath.
func InjectFile(fragletPath string, config *Config) error {
	if config == nil {
		return fmt.Errorf("injection config is required")
	}
	if config.CodePath == "" {
		return fmt.Errorf("injection config must specify codePath")
	}

	// Capture file mode BEFORE any operations
	var mode os.FileMode = 0644
	if info, err := os.Stat(config.CodePath); err == nil {
		mode = info.Mode()
	}

	// Read fraglet content
	fragletData, err := os.ReadFile(fragletPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No fraglet to inject
		}
		return fmt.Errorf("failed to read fraglet file: %w", err)
	}

	// Read target file (create empty if doesn't exist)
	targetData, err := os.ReadFile(config.CodePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read target file: %w", err)
	}
	if os.IsNotExist(err) {
		targetData = []byte{}
	}

	// Inject using string injector
	rendered, err := InjectString(string(targetData), string(fragletData), config)
	if err != nil {
		return fmt.Errorf("injection failed: %w", err)
	}

	// Write result
	if err := os.WriteFile(config.CodePath, []byte(rendered), mode); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	// Ensure file mode is preserved (os.WriteFile may not preserve all mode bits)
	if err := os.Chmod(config.CodePath, mode); err != nil {
		return fmt.Errorf("failed to set file mode: %w", err)
	}

	return nil
}

// injectWithMatchRegion replaces region between match_start and match_end with fraglet code
func injectWithMatchRegion(template string, fragletCode string, matchStart string, matchEnd string) (string, error) {
	lines := strings.Split(template, "\n")
	var result []string
	inRegion := false
	startIndent := ""
	foundStart := false

	for _, line := range lines {
		if !foundStart && strings.Contains(line, matchStart) {
			startIndent = extractIndentation(line)
			fragletLines := strings.Split(fragletCode, "\n")
			for _, fragLine := range fragletLines {
				result = append(result, startIndent+fragLine)
			}
			foundStart = true

			// If matchEnd also on this line, stay out of region
			if strings.Contains(line, matchEnd) {
				inRegion = false
			} else {
				inRegion = true
			}
			continue
		}

		if inRegion {
			if strings.Contains(line, matchEnd) {
				inRegion = false
			}
			continue
		}

		result = append(result, line)
	}

	if !foundStart {
		return "", fmt.Errorf("match_start not found in template: %q", matchStart)
	}

	if inRegion {
		return "", fmt.Errorf("match_end not found in template: %q", matchEnd)
	}

	return strings.Join(result, "\n"), nil
}

// extractIndentation extracts leading whitespace from a line
func extractIndentation(line string) string {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}
