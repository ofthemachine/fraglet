package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ofthemachine/fraglet/entrypoint/internal/executor"
	"github.com/ofthemachine/fraglet/entrypoint/internal/fraglet"
	fragletpkg "github.com/ofthemachine/fraglet/pkg/fraglet"
)

func main() {
	// Load configuration (respects FRAGLET_CONFIG envvar)
	cfg, err := fragletpkg.LoadEntrypointConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "usage":
			fmt.Print(generateUsage(cfg))
			return
		case "guide":
			showDocumentation(cfg, cfg.Guide, "")
			return
		}
	}

	// Process fraglet injection
	fragletMgr := fraglet.NewManager(cfg)
	if err := fragletMgr.Process(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Execute code or pass through args
	exec := executor.NewExecutor(cfg)

	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	exitCode, err := exec.ExecuteWithArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCode)
	}
	os.Exit(exitCode)
}

func showDocumentation(cfg *fragletpkg.EntrypointConfig, configuredPath, notFoundMsg string) {
	// Try configured path first
	if data, err := os.ReadFile(configuredPath); err == nil {
		fmt.Print(string(data))
		return
	}

	// Try in code directory using the filename from configured path
	filename := filepath.Base(configuredPath)
	codeDir := filepath.Dir(cfg.Injection.CodePath)
	codePath := filepath.Join(codeDir, filename)
	if data, err := os.ReadFile(codePath); err == nil {
		fmt.Print(string(data))
		return
	}

	if notFoundMsg != "" {
		fmt.Println(notFoundMsg)
	}
}

const usageTemplate = "# Container Usage\n\n" +
	"## Mount and Run\n\n" +
	"Mount your fraglet code to {{.FragletTempPath}} and run:\n\n" +
	"```bash\n" +
	"docker run --rm -v /path/to/fraglet:{{.FragletTempPath}}:ro <container>\n" +
	"```\n\n" +
	"## Code Injection\n\n" +
	"Code will be injected into `{{.CodePath}}` at the marker: {{.MarkerDisplay}}\n\n" +
	"## Execution\n\n" +
	"After injection, the container executes: `{{.ExecutionPath}}`\n\n" +
	"## Example\n\n" +
	"Here's a functional example using the existing code:\n\n" +
	"```bash\n" +
	"# Read the existing code\n" +
	"cat > /tmp/fraglet.sh << 'EOF'\n" +
	"{{.ExampleCode}}\n" +
	"EOF\n\n" +
	"# Mount and run\n" +
	"docker run --rm -v /tmp/fraglet.sh:{{.FragletTempPath}}:ro <container>\n" +
	"```\n\n" +
	"## Documentation\n\n" +
	"- `usage` - Container usage (this document)\n" +
	"- `guide` - Authoring guide for writing fraglets\n"

type usageData struct {
	FragletTempPath string
	CodePath        string
	MarkerDisplay   string
	ExecutionPath   string
	ExampleCode     string
}

func generateUsage(cfg *fragletpkg.EntrypointConfig) string {
	// Determine marker display
	markerDisplay := markerDisplay(cfg.Injection)

	// Determine execution path
	execPath := cfg.Execution.Path
	if execPath == "" {
		execPath = "<command from args>"
	}

	// Normalize fraglet temp path to absolute
	fragletPath := normalizePath(cfg.FragletTempPath)

	// Extract example code from target file (the code at the marker location)
	exampleCode := extractExampleCode(cfg.Injection)

	data := usageData{
		FragletTempPath: fragletPath,
		CodePath:        cfg.Injection.CodePath,
		MarkerDisplay:   markerDisplay,
		ExecutionPath:   execPath,
		ExampleCode:     exampleCode,
	}

	tmpl, err := template.New("usage").Parse(usageTemplate)
	if err != nil {
		return fmt.Sprintf("Error generating usage: %v\n", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error generating usage: %v\n", err)
	}

	return buf.String()
}

func markerDisplay(inj fragletpkg.InjectionConfig) string {
	if inj.Match != "" {
		return fmt.Sprintf("`%s`", inj.Match)
	}
	if inj.MatchStart != "" && inj.MatchEnd != "" {
		return fmt.Sprintf("`%s` ... `%s`", inj.MatchStart, inj.MatchEnd)
	}
	return "<no marker configured>"
}

func normalizePath(path string) string {
	// If path is already absolute, return as-is
	if strings.HasPrefix(path, "/") {
		return path
	}
	// Otherwise, make it absolute by prepending /
	return "/" + path
}

func extractExampleCode(inj fragletpkg.InjectionConfig) string {
	// Read the target file to extract the code at the marker location
	data, err := os.ReadFile(inj.CodePath)
	if err != nil {
		// Return a simple placeholder if target file doesn't exist
		return "echo 'Hello, World!'"
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var exampleLines []string
	found := false

	if inj.Match != "" {
		// Find the line containing the marker and extract it
		for _, line := range lines {
			if strings.Contains(line, inj.Match) {
				// Return the line as-is (this is what would be replaced)
				exampleLines = append(exampleLines, line)
				found = true
				break
			}
		}
		// If marker not found, file may have been modified by previous injection
		// In this case, we can't determine the original marker line, so use placeholder
	} else if inj.MatchStart != "" && inj.MatchEnd != "" {
		// Extract lines between match_start and match_end
		inRegion := false
		for _, line := range lines {
			if strings.Contains(line, inj.MatchStart) {
				inRegion = true
				// Include the start line
				exampleLines = append(exampleLines, line)
				found = true
				// If matchEnd is also on this line, we're done
				if strings.Contains(line, inj.MatchEnd) {
					break
				}
				continue
			}
			if inRegion {
				exampleLines = append(exampleLines, line)
				if strings.Contains(line, inj.MatchEnd) {
					break
				}
			}
		}
	}

	if !found || len(exampleLines) == 0 {
		return "echo 'Hello, World!'"
	}

	example := strings.Join(exampleLines, "\n")
	// Trim trailing newlines
	example = strings.TrimRight(example, "\n")

	return example
}
