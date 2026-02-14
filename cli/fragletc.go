package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

const defaultFragletPath = "/FRAGLET"

// envListFlag implements flag.Value for repeatable -e flags.
type envListFlag []string

func (e *envListFlag) String() string { return strings.Join(*e, ",") }
func (e *envListFlag) Set(val string) error {
	*e = append(*e, val)
	return nil
}

func main() {
	// Subcommands are checked before flag parsing
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

	// Flags
	veinSpec := flag.String("vein", "", "Vein name with optional mode (e.g., python, python:main)")
	image := flag.String("image", "", "Container image to use directly")
	fragletPath := flag.String("fraglet-path", defaultFragletPath, "Path where code is mounted in container")
	mode := flag.String("mode", "", "Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)")
	inlineCode := flag.String("c", "", "Program passed in as string (like python -c)")
	var envFlags envListFlag
	flag.Var(&envFlags, "e", "Environment variable to forward (repeatable, e.g. -e FOO -e BAR=val)")

	// Short forms
	flag.StringVar(veinSpec, "v", "", "Vein name with optional mode (short form)")
	flag.StringVar(image, "i", "", "Container image (short form)")
	flag.StringVar(fragletPath, "p", defaultFragletPath, "Path where code is mounted (short form)")
	flag.StringVar(mode, "m", "", "Fraglet mode (short form)")
	flag.StringVar(inlineCode, "code", "", "Program passed in as string (like python -c)")

	flag.Parse()

	// Positional: [script-file] [script-args...]
	args := flag.Args()
	var scriptFile string
	var scriptArgs []string
	if len(args) > 0 {
		scriptFile = args[0]
		scriptArgs = args[1:]
	}

	// --- Resolve vein + mode ---
	veinName, finalMode := resolveVeinAndMode(*veinSpec, *mode, *image, scriptFile)

	// Validate mutual exclusion early (before code resolution)
	if *image != "" && veinName != "" {
		fatal("Error: cannot specify both --image and --vein")
	}

	// --- Resolve code ---
	code := resolveCode(*inlineCode, scriptFile)

	// --- Resolve container + fraglet mount path ---
	containerImage, fragletMountPath := resolveContainer(veinName, *image, *fragletPath)

	// --- Build env vars ---
	envVars := buildEnvVars(finalMode, envFlags)

	// --- Write temp file, build spec, execute ---
	tmpFile, cleanup, err := writeTempFile(code)
	if err != nil {
		fatal("Error creating temp file: %v", err)
	}
	defer cleanup()

	r := runner.NewRunner(containerImage, "")
	spec := runner.RunSpec{
		Container:   containerImage,
		Env:         envVars,
		Args:        scriptArgs,
		StdinReader: os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: fragletMountPath,
				ReadOnly:      true,
			},
		},
	}

	result, err := r.Run(context.Background(), spec)
	if err != nil {
		fatal("Execution failed: %v", err)
	}
	os.Exit(result.ExitCode)
}

// resolveVeinAndMode determines the vein name and mode from flags and file extension.
// Exits on validation errors.
func resolveVeinAndMode(veinSpec, modeFlag, image, scriptFile string) (veinName, mode string) {
	if veinSpec != "" {
		var parsedMode string
		var err error
		veinName, parsedMode, err = parseVeinSpec(veinSpec)
		if err != nil {
			fatal("Error: %v", err)
		}
		if modeFlag != "" && parsedMode != "" {
			fatal("Error: mode specified both in --vein and --mode flags")
		}
		if parsedMode != "" {
			mode = parsedMode
		} else {
			mode = modeFlag
		}
		return
	}

	mode = modeFlag

	// Infer vein from file extension when no --vein and no --image
	if scriptFile != "" && image == "" {
		registry, err := loadVeinRegistry()
		if err != nil {
			fatal("Error loading veins: %v", err)
		}
		extMap := vein.NewExtensionMap(registry)
		veinName, err = extMap.VeinForFile(scriptFile)
		if err != nil {
			fatal("Error: %v", err)
		}
	}

	return
}

// resolveCode reads code from -c flag or script file. Exits if no source is provided.
func resolveCode(inlineCode, scriptFile string) string {
	if inlineCode != "" {
		return inlineCode
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			fatal("Error reading file %s: %v", scriptFile, err)
		}
		return stripShebang(string(data))
	}
	fatal("Error: no code source provided. Use a script file or -c flag.")
	return "" // unreachable
}

// resolveContainer determines the container image and fraglet mount path.
// Exits on validation errors.
func resolveContainer(veinName, image, fragletPath string) (containerImage, mountPath string) {
	if veinName != "" {
		registry, err := loadVeinRegistry()
		if err != nil {
			fatal("Error loading veins: %v", err)
		}
		v, ok := registry.Get(veinName)
		if !ok {
			fatal("Error: vein not found: %s", veinName)
		}
		return v.ContainerImage(), defaultFragletPath
	}

	if image != "" {
		return image, fragletPath
	}

	fatal("Error: no container target. Specify --vein or --image.")
	return "", "" // unreachable
}

// buildEnvVars constructs the environment variable list from mode and -e flags.
// For -e entries: "KEY=VALUE" passes through as-is, "KEY" looks up the host env.
func buildEnvVars(mode string, envFlags envListFlag) []string {
	var envVars []string
	if mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", mode))
	}
	for _, entry := range envFlags {
		if strings.Contains(entry, "=") {
			// Explicit KEY=VALUE
			envVars = append(envVars, entry)
		} else {
			// Forward host env var by name
			if val, ok := os.LookupEnv(entry); ok {
				envVars = append(envVars, entry+"="+val)
			}
			// Silently skip vars not set on the host — no leaking, no errors
		}
	}
	return envVars
}

// fatal prints to stderr and exits.
func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// parseVeinSpec parses "vein" or "vein:mode" syntax.
func parseVeinSpec(spec string) (veinName, mode string, err error) {
	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 1:
		return parts[0], "", nil
	case 2:
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid vein spec format: %s (expected 'vein' or 'vein:mode')", spec)
	}
}

// loadVeinRegistry loads veins using the auto-loading mechanism.
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

	registry, err := loadVeinRegistry()
	if err != nil {
		fatal("Error loading veins: %v", err)
	}

	var veinsToRefresh []*vein.Vein

	if *all {
		for _, name := range registry.List() {
			if v, ok := registry.Get(name); ok {
				veinsToRefresh = append(veinsToRefresh, v)
			}
		}
	} else if len(args) > 0 {
		for _, name := range args {
			v, ok := registry.Get(name)
			if !ok {
				fatal("Error: vein not found: %s", name)
			}
			veinsToRefresh = append(veinsToRefresh, v)
		}
	} else {
		refreshFlags.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	platform := "linux/amd64"
	failed := false

	for _, v := range veinsToRefresh {
		img := v.ContainerImage()
		fmt.Printf("Pulling %s (%s)...\n", v.Name, img)
		cmd := exec.CommandContext(ctx, "docker", "pull", "--platform", platform, img)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to pull %s: %v\n", img, err)
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

	registry, err := loadVeinRegistry()
	if err != nil {
		fatal("Error loading veins: %v", err)
	}

	v, ok := registry.Get(veinName)
	if !ok {
		fatal("Error: vein not found: %s", veinName)
	}

	var envVars []string
	if *mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_CONFIG=/fraglet-%s.yml", *mode))
	}

	img := v.ContainerImage()
	r := runner.NewRunner(img, "")
	spec := runner.RunSpec{
		Container: img,
		Env:       envVars,
		Args:      []string{"guide"},
	}

	result, err := r.Run(context.Background(), spec)
	if err != nil {
		fatal("Error running guide: %v", err)
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
  -c, --code string
        Program passed in as string (like python -c)
  -e string
        Environment variable to forward into container (repeatable)
        Use -e FOO to forward host value, -e FOO=bar for explicit value
  -p, --fraglet-path string
        Path where code is mounted in container (default: /FRAGLET)
  -m, --mode string
        Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)

Positional:
  script-file   Path to code file (required if -c not set)
  script-args   Arguments passed to the script inside container

Stdin:
  Stdin is always forwarded to the program inside the container.
  Use pipes to send data: echo "hello" | fragletc --vein=python script.py

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

  # Inline code (like python -c)
  fragletc --vein=python -c 'print("hello")'

  # Pipe data to script
  echo "hello world" | fragletc --vein=python script.py
  cat data.csv | ./process.py --format=json

  # Forward environment variables
  fragletc --vein=python -e DATABASE_URL -e DEBUG=1 script.py

  # In a shebang
  #!/usr/bin/env -S fragletc --vein=python -e API_KEY

  # Vein with mode (via vein:mode syntax)
  fragletc --vein=c:main script.c

  # Direct container image with mode
  fragletc --image=my-registry/python:latest --mode=main script.py

  # As shebang (script contains: #!/usr/bin/env -S fragletc --vein=python)
  ./script.py arg1 arg2

  # Refresh veins
  fragletc refresh ada
  fragletc refresh --all

  # Show guides
  fragletc guide ada
  fragletc guide python
`)
}
