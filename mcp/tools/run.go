package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

var RunTool *mcp.Tool

func init() {
	// Load veins (checks FRAGLET_VEINS_PATH first, then embedded)
	registry, err := vein.LoadAuto(embed.LoadEmbeddedVeins)

	if err != nil {
		// If veins can't be loaded, use placeholder description
		RunTool = &mcp.Tool{
			Name: "run",
			Description: "Execute code snippets for code-based reasoning, leveraging the best language and ecosystem for each task. " +
				"Run small code fragments in isolated sandboxes that may include rich libraries, frameworks, and domain-specific tools. " +
				"Each environment provides not just the language, but potentially fluent interfaces for complex systems interactions, " +
				"data processing libraries, APIs, and more. Use this to explore statistical reasoning, probabilities, mathematical computation, " +
				"physics simulations, data analysis, and other problem domains best solved with code. " +
				"IMPORTANT: Before writing code, use the 'language_help' tool to get the authoring guide for your chosen language—those guides already include everything you need, so do not hunt through repos or veins. " +
				"Each language has specific requirements about code format (e.g., complete programs vs. code fragments, required structure, etc.) " +
				"that you must follow.",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
		}
		return
	}

	veins := registry.List()
	RunTool = &mcp.Tool{
		Name: "run",
		Description: fmt.Sprintf("Execute code snippets for code-based reasoning, leveraging the best language and ecosystem for each task. "+
			"Run small code fragments in isolated sandboxes that may include rich libraries, frameworks, and domain-specific tools. "+
			"Each environment provides not just the language, but potentially fluent interfaces for complex systems interactions, "+
			"data processing libraries, APIs, and more. Use this to explore statistical reasoning, probabilities, mathematical computation, "+
			"physics simulations, data analysis, and other problem domains best solved with code. "+
			"Supported languages: %s. "+
			"IMPORTANT: Before writing code, use the 'language_help' tool to get the authoring guide for your chosen language—those guides already include everything you need, so do not hunt through repos or veins. "+
			"Each language has specific requirements about code format (e.g., complete programs vs. code fragments, required structure, etc.) "+
			"that you must follow. Use this for quick code invocations to test hypotheses, calculate values, analyze data, or prototype solutions.",
			strings.Join(veins, ", ")),
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
	// Load veins
	registry, err := vein.LoadAuto(embed.LoadEmbeddedVeins)
	if err != nil {
		return nil, RunOutput{}, fmt.Errorf("failed to load veins: %w", err)
	}

	// Get vein by name
	v, ok := registry.Get(input.Lang)
	if !ok {
		return nil, RunOutput{}, fmt.Errorf("vein not found: %s", input.Lang)
	}

	// Write code to temp file
	tmpFile, cleanup, err := writeTempFile(input.Code)
	if err != nil {
		return nil, RunOutput{}, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer cleanup()

	// Create runner
	r := runner.NewRunner(v.Container, "")

	// Execute with volume mount
	spec := runner.RunSpec{
		Container: v.Container,
		Args:      nil, // No script args for MCP
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: "/FRAGLET",
				ReadOnly:      true,
			},
		},
	}

	result, err := r.Run(ctx, spec)
	if err != nil {
		return nil, RunOutput{}, fmt.Errorf("execution failed: %w", err)
	}

	// Format output for better rendering
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

	// Format code block
	codeFence := fenceForCodeBlock(input.Code)
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

func writeTempFile(content string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "fraglet-*")
	if err != nil {
		return "", nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0644)

	absPath, _ := filepath.Abs(tmpFile.Name())
	cleanup := func() { os.Remove(absPath) }

	return absPath, cleanup, nil
}
