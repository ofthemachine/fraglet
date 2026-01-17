package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

const defaultFragletPath = "/FRAGLET"

func main() {
	// Check for subcommands before parsing flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "refresh":
			handleRefresh()
			return
		case "guide":
			handleGuide()
			return
		}
	}

	flag.Usage = usage

	// Define flags
	veinSpec := flag.String("vein", "", "Vein name with optional mode (e.g., python, python:main)")
	image := flag.String("image", "", "Container image to use directly (e.g., my-registry/python:latest)")
	fragletPath := flag.String("fraglet-path", defaultFragletPath, "Path where code is mounted in container (default: /FRAGLET)")
	mode := flag.String("mode", "", "Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)")

	// Short forms
	flag.StringVar(veinSpec, "v", "", "Vein name with optional mode (short form)")
	flag.StringVar(image, "i", "", "Container image (short form)")
	flag.StringVar(fragletPath, "p", defaultFragletPath, "Path where code is mounted in container (short form)")
	flag.StringVar(mode, "m", "", "Fraglet mode (short form)")

	flag.Parse()

	// Handle positional arguments
	args := flag.Args()
	var scriptFile string
	var scriptArgs []string

	if len(args) > 0 {
		if args[0] == "-" {
			// Explicit stdin, remaining args are script args
			scriptArgs = args[1:]
		} else {
			// First arg is script file, remaining are script args
			scriptFile = args[0]
			scriptArgs = args[1:]
		}
	}

	// Determine vein and mode
	var veinName, finalMode string
	var err error

	if *veinSpec != "" {
		// Parse vein:mode syntax
		var parsedMode string
		veinName, parsedMode, err = parseVeinSpec(*veinSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		// Use mode from flag if provided, otherwise use parsed mode
		if *mode != "" {
			if parsedMode != "" {
				fmt.Fprintf(os.Stderr, "Error: mode specified both in --vein and --mode flags\n")
				os.Exit(1)
			}
			finalMode = *mode
		} else {
			finalMode = parsedMode
		}
	} else if *mode != "" {
		// Mode specified without vein (only valid with --image)
		finalMode = *mode
	}

	if scriptFile != "" && *image == "" && veinName == "" {
		// Try to infer from extension (only if --image is not specified)
		registry, regErr := loadVeinRegistry()
		if regErr != nil {
			fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", regErr)
			os.Exit(1)
		}
		extMap := vein.NewExtensionMap(registry)
		veinName, err = extMap.VeinForFile(scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else if *image == "" && veinName == "" {
		// No vein, no image, no file - error
		usage()
		os.Exit(1)
	}

	// Validate: cannot specify both image and vein
	if *image != "" && veinName != "" {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both --image and --vein\n")
		os.Exit(1)
	}

	var code string

	// Read code from file if provided, otherwise from STDIN
	if scriptFile != "" {
		codeBytes, err := os.ReadFile(scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", scriptFile, err)
			os.Exit(1)
		}
		code = string(codeBytes)
		// Strip shebang if present
		code = stripShebang(code)
	} else {
		// Read from STDIN
		codeBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from STDIN: %v\n", err)
			os.Exit(1)
		}
		code = string(codeBytes)
	}

	ctx := context.Background()
	var result *runner.RunResult

	if veinName != "" {
		// Use vein
		registry, err := loadVeinRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
			os.Exit(1)
		}

		v, ok := registry.Get(veinName)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: vein not found: %s\n", veinName)
			os.Exit(1)
		}

		// Build environment variables for mode
		var envVars []string
		if finalMode != "" {
			// Mode convention: /fraglet-{mode}.yml or /fraglet-{mode}.yaml
			envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", finalMode))
		}

		// Write code to temp file
		tmpFile, cleanup, err := writeTempFile(code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		// Create runner
		r := runner.NewRunner(v.Container, "")

		// Execute with volume mount at fragletPath
		spec := runner.RunSpec{
			Container: v.Container,
			Env:       envVars,
			Args:      scriptArgs,
			Volumes: []runner.VolumeMount{
				{
					HostPath:      tmpFile,
					ContainerPath: defaultFragletPath,
					ReadOnly:      true,
				},
			},
		}

		runResult, runErr := r.Run(ctx, spec)
		result = &runResult
		err = runErr
	} else {
		// Use direct container image
		containerImage := *image
		finalFragletPath := *fragletPath

		// Build environment variables for mode
		var envVars []string
		if finalMode != "" {
			// Mode convention: /fraglet-{mode}.yml or /fraglet-{mode}.yaml
			envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", finalMode))
		}

		// Write code to temp file
		tmpFile, cleanup, err := writeTempFile(code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		// Create runner
		r := runner.NewRunner(containerImage, "")

		// Execute with volume mount at fragletPath
		spec := runner.RunSpec{
			Container: containerImage,
			Env:       envVars,
			Args:      scriptArgs,
			Volumes: []runner.VolumeMount{
				{
					HostPath:      tmpFile,
					ContainerPath: finalFragletPath,
					ReadOnly:      true,
				},
			},
		}

		runResult, runErr := r.Run(ctx, spec)
		result = &runResult
		err = runErr
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	// Exit with the same code as the execution
	os.Exit(result.ExitCode)
}

// parseVeinSpec parses "vein" or "vein:mode" syntax
func parseVeinSpec(spec string) (veinName, mode string, err error) {
	parts := strings.Split(spec, ":")
	if len(parts) == 1 {
		return parts[0], "", nil
	}
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("invalid vein spec format: %s (expected 'vein' or 'vein:mode')", spec)
}

// loadVeinRegistry loads veins using the auto-loading mechanism
func loadVeinRegistry() (*vein.VeinRegistry, error) {
	return vein.LoadAuto(embed.LoadEmbeddedVeins)
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

func stripShebang(code string) string {
	if strings.HasPrefix(code, "#!") {
		if idx := strings.Index(code, "\n"); idx != -1 {
			return code[idx+1:]
		}
	}
	return code
}

func handleRefresh() {
	refreshFlags := flag.NewFlagSet("refresh", flag.ExitOnError)
	all := refreshFlags.Bool("all", false, "Refresh all veins")
	refreshFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: fragletc refresh [options] [vein-name]

Refresh (pull) container images for veins.

Options:
  --all    Refresh all veins

Examples:
  fragletc refresh ada              # Refresh ada vein
  fragletc refresh --all            # Refresh all veins

The command respects FRAGLET_VEINS_PATH environment variable for custom veins.
`)
	}

	refreshFlags.Parse(os.Args[2:])
	args := refreshFlags.Args()

	// Load vein registry (respects FRAGLET_VEINS_PATH)
	registry, err := loadVeinRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
		os.Exit(1)
	}

	var veinsToRefresh []*vein.Vein

	if *all {
		// Refresh all veins
		veinNames := registry.List()
		for _, name := range veinNames {
			v, ok := registry.Get(name)
			if !ok {
				continue
			}
			veinsToRefresh = append(veinsToRefresh, v)
		}
	} else if len(args) > 0 {
		// Refresh specific vein(s)
		for _, name := range args {
			v, ok := registry.Get(name)
			if !ok {
				fmt.Fprintf(os.Stderr, "Error: vein not found: %s\n", name)
				os.Exit(1)
			}
			veinsToRefresh = append(veinsToRefresh, v)
		}
	} else {
		refreshFlags.Usage()
		os.Exit(1)
	}

	// Pull images
	ctx := context.Background()
	platform := "linux/amd64"
	failed := false

	for _, v := range veinsToRefresh {
		fmt.Printf("Pulling %s (%s)...\n", v.Name, v.Container)
		cmd := exec.CommandContext(ctx, "docker", "pull", "--platform", platform, v.Container)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to pull %s: %v\n", v.Container, err)
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
}

func handleGuide() {
	guideFlags := flag.NewFlagSet("guide", flag.ExitOnError)
	mode := guideFlags.String("mode", "", "Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)")
	guideFlags.StringVar(mode, "m", "", "Fraglet mode (short form)")
	guideFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: fragletc guide [options] <vein-name>

Show the fraglet guide for a specific vein.

Options:
  -m, --mode string
        Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)

Examples:
  fragletc guide ada                    # Show guide for ada vein
  fragletc guide python                  # Show guide for python vein
  fragletc guide ada --mode main         # Show guide for ada vein with main mode

The command respects FRAGLET_VEINS_PATH environment variable for custom veins.
`)
	}

	guideFlags.Parse(os.Args[2:])
	args := guideFlags.Args()

	if len(args) == 0 {
		guideFlags.Usage()
		os.Exit(1)
	}

	veinName := args[0]

	// Load vein registry (respects FRAGLET_VEINS_PATH)
	registry, err := loadVeinRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
		os.Exit(1)
	}

	v, ok := registry.Get(veinName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: vein not found: %s\n", veinName)
		os.Exit(1)
	}

	// Build environment variables for mode
	var envVars []string
	if *mode != "" {
		// Mode convention: /fraglet-{mode}.yml or /fraglet-{mode}.yaml
		envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", *mode))
	}

	// Run container with "guide" as first argument
	ctx := context.Background()
	r := runner.NewRunner(v.Container, "")
	spec := runner.RunSpec{
		Container: v.Container,
		Env:       envVars,
		Args:      []string{"guide"},
	}

	result, err := r.Run(ctx, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running guide: %v\n", err)
		os.Exit(1)
	}

	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}

	os.Exit(result.ExitCode)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: fragletc [flags] [script-file] [script-args...]
       fragletc refresh [options] [vein-name]

Execute fraglet code in a container using either --vein or --image.

Flags:
  -v, --vein string
        Vein name with optional mode (e.g., python, python:main)
  -i, --image string
        Container image to use directly (e.g., my-registry/python:latest)
  -p, --fraglet-path string
        Path where code is mounted in container (default: /FRAGLET)
  -m, --mode string
        Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)

Positional:
  script-file   Path to code file, or "-" for stdin
  script-args   Arguments passed to the script inside container

Subcommands:
  refresh       Refresh (pull) container images for veins
                Use "fragletc refresh --help" for details
  guide         Show fraglet guide for a vein
                Use "fragletc guide --help" for details

Examples:
  # Infer vein from extension
  fragletc script.py
  fragletc script.py arg1 arg2

  # Explicit vein
  fragletc --vein=python script.py

  # Vein with mode (via vein:mode syntax)
  fragletc --vein=c:main script.c

  # Vein with mode (via --mode flag)
  fragletc --vein=python --mode=main script.py

  # Direct container image with mode
  fragletc --image=my-registry/python:latest --mode=main script.py

  # As shebang (script contains: #!/usr/bin/fragletc)
  ./script.py arg1 arg2

  # Refresh veins
  fragletc refresh ada
  fragletc refresh --all

  # Show guides
  fragletc guide ada
  fragletc guide python
`)
}
