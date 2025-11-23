package language

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GetAgentHelp retrieves the guide documentation from a language container
// by running: docker run <container> guide
func GetAgentHelp(ctx context.Context, lang string) (string, error) {
	config, ok := GetLanguage(lang)
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", lang)
	}

	// Run docker command to get guide
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", config.Container, "guide")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get guide for %s: %w (stderr: %s)", lang, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

