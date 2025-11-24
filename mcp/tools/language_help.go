package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

var LanguageHelpTool = &mcp.Tool{
	Name:        "language_help",
	Description: "Get the authoring guide for a language environment. Use this to learn syntax, code patterns, available libraries, frameworks, and domain-specific tools for writing code fragments. Each language container may include rich ecosystems—libraries for complex systems interactions, data processing, APIs, and more. Essential for code-based reasoning—helps you discover and leverage the full capabilities of each environment for statistical analysis, mathematical computation, physics simulations, probability calculations, and other problem domains best explored with code.",
	Annotations: &mcp.ToolAnnotations{
		ReadOnlyHint: true,
	},
}

type LanguageHelpInput struct {
	Lang string `json:"lang" jsonschema:"the language to get the authoring guide for (e.g., 'python', 'r', 'lisp')"`
}

type LanguageHelpOutput struct {
	Help string `json:"help" jsonschema:"markdown guide documentation about the language (authoring guide from the container)"`
}

func LanguageHelp(ctx context.Context, req *mcp.CallToolRequest, input LanguageHelpInput) (
	*mcp.CallToolResult,
	LanguageHelpOutput,
	error,
) {
	// Create fraglet environment to get envelope
	envelopesDir := findEnvelopesDir()
	env, err := fraglet.NewFragletEnvironment(envelopesDir)
	if err != nil {
		return nil, LanguageHelpOutput{}, fmt.Errorf("failed to init environment: %w", err)
	}

	envelope, ok := env.GetRegistry().GetEnvelope(input.Lang)
	if !ok {
		return nil, LanguageHelpOutput{}, fmt.Errorf("unsupported language: %s", input.Lang)
	}

	// Run docker command to get guide
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", envelope.Container, "guide")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, LanguageHelpOutput{}, fmt.Errorf("failed to get guide for %s: %w (stderr: %s)", input.Lang, err, stderr.String())
	}

	help := strings.TrimSpace(stdout.String())

	// Return formatted TextContent for better rendering in chat
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: help,
			},
		},
	}, LanguageHelpOutput{Help: help}, nil
}
