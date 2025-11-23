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
	Description: "Stuck authoring a fragment? Use this to learn about any supported language—covers syntax, runner expectations, and fragment authoring tips straight from it's execution environment.",
	Annotations: &mcp.ToolAnnotations{
		ReadOnlyHint: true,
	},
}

type LanguageHelpInput struct {
	Lang string `json:"lang" jsonschema:"the language to get help for (e.g., 'python')"`
}

type LanguageHelpOutput struct {
	Help string `json:"help" jsonschema:"markdown documentation about the language"`
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
