package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/save"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

var RunTool *mcp.Tool

const runToolDescriptionBase = "Execute code snippets for code-based reasoning, leveraging the best language and ecosystem for each task. " +
	"Run small code fragments in isolated sandboxes that may include rich libraries, frameworks, and domain-specific tools. " +
	"Each environment provides not just the language, but potentially fluent interfaces for complex systems interactions, " +
	"data processing libraries, APIs, and more. Use this to explore statistical reasoning, probabilities, mathematical computation, " +
	"physics simulations, data analysis, and other problem domains best solved with code. " +
	"IMPORTANT: Before writing code, use the 'language_help' tool to get the authoring guide for your chosen language—those guides already include everything you need, so do not hunt through repos or veins. " +
	"Each language has specific requirements about code format (e.g., complete programs vs. code fragments, required structure, etc.) " +
	"that you must follow. " +
	"A fraglet is an executable code file that runs in a containerized sandbox. When writing fraglets to disk, use the shebang #!/usr/bin/env -S fragletc --vein=<lang> where <lang> matches the lang parameter. Files with this shebang are directly executable via ./filename. " +
	"Optional: mode and annotations (key:value tokens, e.g. determinism:deterministic, math:number-theory)."

func init() {
	registry, err := vein.LoadAuto(embed.LoadEmbeddedVeins)
	if err != nil {
		RunTool = &mcp.Tool{
			Name:        "run",
			Description: runToolDescriptionBase,
			Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		}
		return
	}
	RunTool = &mcp.Tool{
		Name: "run",
		Description: runToolDescriptionBase +
			"Supported languages: " + strings.Join(registry.List(), ", ") + ". " +
			"Runs are limited to 60s by default; pass timeout_seconds to override. ",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}
}

const DefaultRunTimeout = 60 * time.Second

type RunInput struct {
	Lang           string            `json:"lang" jsonschema:"the language (vein) to run the code in"`
	Code           string            `json:"code" jsonschema:"the code to run"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty" jsonschema:"max execution time in seconds; default 60, 0 means use default"`
	Mode           string            `json:"mode,omitempty" jsonschema:"optional mode; when provided, uses that execution mode for the container"`
	Annotations    []string          `json:"annotations,omitempty" jsonschema:"optional key:value tokens (e.g. determinism:deterministic, math:number-theory)"`
	Params         map[string]string `json:"params,omitempty" jsonschema:"parameters injected as env vars, keyed by alias with optional type prefix (e.g. raw, b64, cb64)"`
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

	// Apply timeout: default 60s, overridable via timeout_seconds (0 = use default)
	timeout := DefaultRunTimeout
	if input.TimeoutSeconds > 0 {
		timeout = time.Duration(input.TimeoutSeconds) * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create runner (ContainerImage applies FRAGLET_VEINS_FORCE_TAG if set)
	img := v.ContainerImage()
	r := runner.NewRunner(img, "")

	// Build env: optional FRAGLET_CONFIG when mode is set
	var envVars []string
	if input.Mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", input.Mode))
	}

	// Parse and resolve params
	var params fraglet.Params
	if len(input.Params) > 0 {
		for alias, value := range input.Params {
			p, err := fraglet.ParseParam(alias + "=" + value)
			if err != nil {
				return nil, RunOutput{}, fmt.Errorf("param %q: %w", alias, err)
			}
			params = append(params, p)
		}
		decls := fraglet.ParseParamDecls(input.Code)
		if len(decls) > 0 {
			var err error
			params, err = params.ResolveAliases(decls)
			if err != nil {
				return nil, RunOutput{}, fmt.Errorf("param resolution: %w", err)
			}
		}
		transportEnv, err := params.ToTransportEnv()
		if err != nil {
			return nil, RunOutput{}, fmt.Errorf("param transport env: %w", err)
		}
		envVars = append(envVars, transportEnv...)
	}

	// Execute with volume mount. Stdin and script args are not passed through the MCP run tool (code-only).
	spec := runner.RunSpec{
		Container: img,
		Env:       envVars,
		Args:      nil,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: "/FRAGLET",
				// Writable: false (default) = read-only mount
			},
		},
	}

	result, err := r.Run(runCtx, spec)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timeoutMsg := fmt.Sprintf("execution timed out after %s", timeout)
			result = runner.RunResult{Stderr: timeoutMsg, ExitCode: 124}
		} else {
			return nil, RunOutput{}, fmt.Errorf("execution failed: %w", err)
		}
	}

	// Persist on success when save path is configured (invisible to agent: no path/hash in response)
	if result.ExitCode == 0 {
		if saveRoot := getRunSavePath(); saveRoot != "" {
			imageWithDigest, _ := vein.ResolveImageDigest(runCtx, img)
			saver := &save.LocalSave{Root: saveRoot}
			_ = saver.Save(runCtx, input.Lang, imageWithDigest, input.Mode, input.Annotations, input.Code)
		}
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
