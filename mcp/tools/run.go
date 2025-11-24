package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

var RunTool *mcp.Tool

// findEnvelopesDir finds the envelopes directory by trying multiple paths
func findEnvelopesDir() string {
	// Try environment variable first
	if dir := os.Getenv("FRAGLET_ENVELOPES_DIR"); dir != "" {
		return dir
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err == nil {
		// Try relative to current working directory
		envelopesPath := filepath.Join(cwd, "envelopes")
		if _, err := os.Stat(envelopesPath); err == nil {
			return envelopesPath
		}

		// Try going up directories to find repo root
		current := cwd
		for i := 0; i < 5; i++ {
			envelopesPath := filepath.Join(current, "envelopes")
			if _, err := os.Stat(envelopesPath); err == nil {
				return envelopesPath
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}

	// Try relative to executable (for installed binaries)
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		// Try going up from bin/ or build/ directories
		for _, base := range []string{execDir, filepath.Dir(execDir), filepath.Dir(filepath.Dir(execDir))} {
			envelopesPath := filepath.Join(base, "envelopes")
			if _, err := os.Stat(envelopesPath); err == nil {
				return envelopesPath
			}
		}
	}

	// Default fallback (relative path)
	return "./envelopes"
}

func init() {
	// Initialize with envelopes directory
	envelopesDir := findEnvelopesDir()
	env, err := fraglet.NewFragletEnvironment(envelopesDir)
	if err != nil {
		// If envelopes can't be loaded, use placeholder description
		RunTool = &mcp.Tool{
			Name: "run",
			Description: "Execute code snippets for code-based reasoning, leveraging the best language and ecosystem for each task. " +
				"Run small code fragments in isolated sandboxes that may include rich libraries, frameworks, and domain-specific tools. " +
				"Each environment provides not just the language, but potentially fluent interfaces for complex systems interactions, " +
				"data processing libraries, APIs, and more. Use this to explore statistical reasoning, probabilities, mathematical computation, " +
				"physics simulations, data analysis, and other problem domains best solved with code. " +
				"IMPORTANT: Before writing code, use the 'language_help' tool to get the authoring guide for your chosen language. " +
				"Each language has specific requirements about code format (e.g., complete programs vs. code fragments, required structure, etc.) " +
				"that you must follow.",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
		}
		return
	}

	envelopes := env.GetRegistry().ListEnvelopes()
	RunTool = &mcp.Tool{
		Name: "run",
		Description: fmt.Sprintf("Execute code snippets for code-based reasoning, leveraging the best language and ecosystem for each task. "+
			"Run small code fragments in isolated sandboxes that may include rich libraries, frameworks, and domain-specific tools. "+
			"Each environment provides not just the language, but potentially fluent interfaces for complex systems interactions, "+
			"data processing libraries, APIs, and more. Use this to explore statistical reasoning, probabilities, mathematical computation, "+
			"physics simulations, data analysis, and other problem domains best solved with code. "+
			"Supported languages: %s. "+
			"IMPORTANT: Before writing code, use the 'language_help' tool to get the authoring guide for your chosen language. "+
			"Each language has specific requirements about code format (e.g., complete programs vs. code fragments, required structure, etc.) "+
			"that you must follow. Use this for quick code invocations to test hypotheses, calculate values, analyze data, or prototype solutions.",
			strings.Join(envelopes, ", ")),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

type RunInput struct {
	Lang string `json:"lang" jsonschema:"the language to run the code in"`
	Code string `json:"code" jsonschema:"the code to run"`
}

type RunOutput struct {
	Stdout   string        `json:"standard_out" jsonschema:"the standard output of the code"`
	Stderr   string        `json:"standard_error" jsonschema:"the standard error of the code"`
	ExitCode int           `json:"exit_code" jsonschema:"the exit code of the code"`
	Duration time.Duration `json:"duration" jsonschema:"the duration of the code execution"`
}

func Run(ctx context.Context, req *mcp.CallToolRequest, input RunInput) (
	*mcp.CallToolResult,
	RunOutput,
	error,
) {
	// Create fraglet environment
	envelopesDir := findEnvelopesDir()
	env, err := fraglet.NewFragletEnvironment(envelopesDir)
	if err != nil {
		return nil, RunOutput{}, fmt.Errorf("failed to init environment: %w", err)
	}

	// Create FragletProc (just the code)
	proc := fraglet.NewFragletProc(input.Code)

	// Execute using envelope name from input.Lang
	result, err := env.Execute(ctx, input.Lang, proc)
	if err != nil {
		return nil, RunOutput{}, fmt.Errorf("execution failed: %w", err)
	}

	// Format output for better rendering - wrap in code blocks for safety
	// Don't assume stdout/stderr are markdown, so we present them as plain text code blocks
	// Use a fence length that's longer than any sequence of backticks in the content
	fenceForCodeBlock := func(content string) string {
		fence := "```"
		for strings.Contains(content, fence) {
			fence += "`"
		}
		return fence
	}

	var contentParts []string

	if result.Stdout != "" {
		fence := fenceForCodeBlock(result.Stdout)
		contentParts = append(contentParts, fmt.Sprintf("**Standard Output:**\n%s\n%s\n%s", fence, result.Stdout, fence))
	}

	if result.Stderr != "" {
		fence := fenceForCodeBlock(result.Stderr)
		contentParts = append(contentParts, fmt.Sprintf("**Standard Error:**\n%s\n%s\n%s", fence, result.Stderr, fence))
	}

	// Add execution metadata
	status := "✅ Success"
	if result.ExitCode != 0 {
		status = fmt.Sprintf("❌ Failed (exit code: %d)", result.ExitCode)
	}
	contentParts = append(contentParts, fmt.Sprintf("**Status:** %s | **Duration:** %s", status, result.Duration.Round(time.Millisecond)))

	// Also handle backticks in the input code
	codeFence := fenceForCodeBlock(input.Code)
	// Format code block with explicit newlines for proper markdown rendering
	// Structure: title, blank line, opening fence+lang on one line, code on subsequent lines, closing fence, blank line, output
	codeBlock := fmt.Sprintf("%s%s\n%s\n%s", codeFence, input.Lang, input.Code, codeFence)
	formattedContent := fmt.Sprintf("**Code executed in `%s`:**\n\n%s\n\n%s",
		input.Lang, codeBlock, strings.Join(contentParts, "\n\n"))

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: formattedContent,
				},
			},
		}, RunOutput{
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
			ExitCode: result.ExitCode,
			Duration: result.Duration,
		}, nil
}
