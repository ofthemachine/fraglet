package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		case "how-to":
			showDocumentation(cfg, cfg.HowTo, "No how-to documentation found in this container.")
			return
		case "agent-help", "help":
			showDocumentation(cfg, cfg.AgentHelp, "No agent-help documentation found in this container.")
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

	fmt.Println(notFoundMsg)
}
