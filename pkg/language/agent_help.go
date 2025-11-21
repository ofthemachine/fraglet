package language

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GetAgentHelp retrieves the agent-help documentation from a language container
// by running: docker run <container> agent-help
func GetAgentHelp(ctx context.Context, lang string) (string, error) {
	config, ok := GetLanguage(lang)
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", lang)
	}

	// Run docker command to get agent-help
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", config.Container, "agent-help")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get agent-help for %s: %w (stderr: %s)", lang, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

