package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/guide"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

var LanguageHelpTool = &mcp.Tool{
	Name:        "language_help",
	Description: "Get the authoring guide for a language environment. Use this to learn syntax, code patterns, available libraries, frameworks, and domain-specific tools for writing code fragments. Each language container may include rich ecosystems—libraries for complex systems interactions, data processing, APIs, and more. Essential for code-based reasoning—helps you discover and leverage the full capabilities of each environment for statistical analysis, mathematical computation, physics simulations, probability calculations, and other problem domains best explored with code. The returned guide applies to both inline code passed to the 'run' tool and standalone fraglet files. Fraglet files use the shebang #!/usr/bin/env -S fragletc --vein=<lang> and are directly executable. Some languages support modes; pass the optional mode parameter to get the guide for that mode, and use the same lang and mode in the run tool.",
	Annotations: &mcp.ToolAnnotations{
		ReadOnlyHint: true,
	},
}

type LanguageHelpInput struct {
	Lang string `json:"lang" jsonschema:"the language (vein) to get the authoring guide for"`
	Mode string `json:"mode,omitempty" jsonschema:"optional mode; when provided, returns the guide for that mode"`
}

type LanguageHelpOutput struct {
	Help string `json:"help" jsonschema:"markdown guide documentation about the language (authoring guide for the context)"`
}

func LanguageHelp(ctx context.Context, req *mcp.CallToolRequest, input LanguageHelpInput) (
	*mcp.CallToolResult,
	LanguageHelpOutput,
	error,
) {
	registry, err := vein.LoadAuto(embed.LoadEmbeddedVeins)
	if err != nil {
		return nil, LanguageHelpOutput{}, fmt.Errorf("failed to load veins: %w", err)
	}

	result, err := guide.Run(ctx, registry, input.Lang, input.Mode)
	if err != nil {
		return nil, LanguageHelpOutput{}, fmt.Errorf("failed to get guide for %s: %w", input.Lang, err)
	}
	if result.ExitCode != 0 {
		return nil, LanguageHelpOutput{}, fmt.Errorf("failed to get guide for %s (exit %d): %s", input.Lang, result.ExitCode, result.Stderr)
	}

	help := strings.TrimSpace(result.Stdout)
	reminder := "\n\n---\nFraglet handles code injection/execution for you. Treat this authoring guide as the single source of truth—no repo spelunking or vein inspection required. If execution fails, iterate from that feedback rather than hunting for config. When writing fraglet files to disk, always use the shebang: #!/usr/bin/env -S fragletc --vein=<lang>"
	if input.Mode != "" {
		reminder += fmt.Sprintf(" For this mode, pass lang=%q and mode=%q to the run tool.", input.Lang, input.Mode)
	}
	reminder += "\n"
	helpWithReminder := help + reminder

	// Return formatted TextContent for better rendering in chat
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: helpWithReminder,
			},
		},
	}, LanguageHelpOutput{Help: helpWithReminder}, nil
}
