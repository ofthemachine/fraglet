package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/mcp/tools"
	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/guide"
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
		case "mcp":
			handleMCP()
			return
		case "version":
			handleVersion()
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
	flag.StringVar(mode, "m", "", "Fraglet mode (short form)")
	flag.StringVar(inlineCode, "code", "", "Program passed in as string (like python -c)")

	// Strip --fraglet-help and extract -p / --param from the full argv (anywhere), then flag.Parse.
	// Params are not registered on the flag set — only this pass sees them. Tokens after "--"
	// are left alone (--fraglet-help, -p, etc. pass through to the program).
	filtered, wantFragletHelp, paramStrs, err := preprocessFragletArgv(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	if err := flag.CommandLine.Parse(filtered); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	// Positional: [script-file] [script-args...]
	args := flag.Args()
	var scriptFile string
	var scriptArgs []string
	if len(args) > 0 {
		scriptFile = args[0]
		scriptArgs = args[1:]
	}

	// --- Handle --fraglet-help (consumed from argv; never forwarded to the script) ---
	if wantFragletHelp {
		handleFragletHelp(scriptFile, *inlineCode)
		return
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

	// --- Parse and resolve params (FRAGLET_PARAM_* transport; entrypoint coerces when present) ---
	var params fraglet.Params
	if len(paramStrs) > 0 {
		for _, pf := range paramStrs {
			p, err := fraglet.ParseParam(pf)
			if err != nil {
				fatal("Error: %v", err)
			}
			params = append(params, p)
		}
		// Resolve aliases via fraglet-meta declarations if code is available
		decls := fraglet.ParseParamDecls(code)
		if len(decls) > 0 {
			var err error
			params, err = params.ResolveAliases(decls)
			if err != nil {
				fatal("Error: %v", err)
			}
		}
		transportEnv, err := params.ToTransportEnv()
		if err != nil {
			fatal("Error: %v", err)
		}
		envVars = append(envVars, transportEnv...)
	}

	// --- Write temp file, build spec, execute ---
	tmpFile, cleanup, err := writeTempFile(code)
	if err != nil {
		fatal("Error creating temp file: %v", err)
	}
	defer cleanup()

	// Only attach stdin when it's a pipe or file; if it's a TTY, leave nil so the runner
	// doesn't use -i and the container exits when the program ends instead of waiting for input.
	var stdinReader io.Reader
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		stdinReader = os.Stdin
	}

	r := runner.NewRunner(containerImage, "")
	spec := runner.RunSpec{
		Container:   containerImage,
		Env:         envVars,
		Args:        scriptArgs,
		StdinReader: stdinReader,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: fragletMountPath,
				// Writable: false (default) = read-only mount
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

// fragletHelpLabel is the name shown for --fraglet-help output: argv[0] basename, or "<inline>".
func fragletHelpLabel(scriptFile string) string {
	if scriptFile == "" {
		return "<inline>"
	}
	return filepath.Base(scriptFile)
}

func handleFragletHelp(scriptFile, inlineCode string) {
	code := inlineCode
	if code == "" && scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			fatal("Error reading file %s: %v", scriptFile, err)
		}
		code = string(data)
	}
	if code == "" {
		fatal("Error: --fraglet-help requires a script file or -c code")
	}

	decls := fraglet.ParseParamDecls(code)
	desc := fraglet.ParseMetaDescription(code)
	label := fragletHelpLabel(scriptFile)

	if desc != "" {
		fmt.Fprintf(os.Stdout, "%s\n\n", desc)
	}

	if len(decls) == 0 {
		fmt.Printf("No parameters declared in %s.\n", label)
		fmt.Fprintf(os.Stdout, "\nAdd param= under fraglet-meta to list names here; optional description= or d= on its own fraglet-meta line.\n")
		return
	}

	fmt.Printf("Parameters for %s:\n", label)
	for _, d := range decls {
		var parts []string
		if d.IsRequired() {
			parts = append(parts, "required")
		} else {
			parts = append(parts, "optional")
		}
		if def, ok := d.Default(); ok {
			parts = append(parts, "default: "+def)
		}
		modStr := strings.Join(parts, ", ")
		fmt.Printf("  %-12s (%s)%s\n", d.Alias, modStr, envVarArrow(d))
	}
	printFragletInvokeHint(label)
}

// printFragletInvokeHint is short user-facing text (shebang-first). Detail lives in fragletc --help.
func printFragletInvokeHint(label string) {
	if label == "<inline>" {
		fmt.Fprintf(os.Stdout, "\nPass: fragletc --vein=<vein> -p name=value ... -c '<code>'\n")
		return
	}
	fmt.Fprintf(os.Stdout, "\nPass: ./%s -p name=value ...  (repeat -p per parameter; see fragletc --help)\n", label)
}

func envVarArrow(d fraglet.ParamDecl) string {
	defaultEnv := strings.ToUpper(d.Alias)
	if d.EnvVar != defaultEnv {
		return fmt.Sprintf("    → %s", d.EnvVar)
	}
	return ""
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

	guideFlags.Parse(reorderArgs(os.Args[2:]))
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

	result, err := guide.Run(context.Background(), registry, veinName, *mode)
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

func handleMCP() {
	mcpFlags := flag.NewFlagSet("mcp", flag.ExitOnError)
	savePath := mcpFlags.String("save", "", "Directory to persist successfully run fraglets (content-addressed); optional")
	mcpFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: fragletc mcp [options]

Start the MCP (Model Context Protocol) server over stdio.

Options:
  --save path   If set, successfully run fraglets are persisted under path (by lang and content hash).
                Use with Cursor, Claude Desktop, or any MCP-compatible client.

Examples:
  fragletc mcp
  fragletc mcp --save=$HOME/.fraglet/store
`)
	}
	_ = mcpFlags.Parse(os.Args[2:])
	if *savePath != "" {
		tools.SetRunSavePath(expandSavePath(*savePath))
	}
	tools.Server.Run(context.Background(), &mcp.StdioTransport{})
}

// expandSavePath expands $VAR and ${VAR} in path, and replaces leading ~ with the user's home directory.
func expandSavePath(path string) string {
	path = os.ExpandEnv(path)
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

const fragletHelpArg = "--fraglet-help"

// preprocessFragletArgv removes --fraglet-help and -p/--param (with values) from argv before
// flag.Parse. Extraction applies to the whole command line except tokens after a bare "--"
// (passthrough). Bundled "-pKEY=value" is accepted only when '=' appears in the suffix so
// flags like -path are not treated as params.
func preprocessFragletArgv(args []string) (filtered []string, wantHelp bool, params []string, err error) {
	passthrough := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if passthrough {
			filtered = append(filtered, a)
			continue
		}
		if a == "--" {
			filtered = append(filtered, a)
			passthrough = true
			continue
		}
		if a == fragletHelpArg {
			wantHelp = true
			continue
		}
		if strings.HasPrefix(a, "--param=") {
			params = append(params, a[len("--param="):])
			continue
		}
		if a == "--param" {
			if i+1 >= len(args) {
				return nil, false, nil, errors.New("--param requires a value")
			}
			i++
			params = append(params, args[i])
			continue
		}
		if strings.HasPrefix(a, "-p=") {
			params = append(params, a[len("-p="):])
			continue
		}
		if a == "-p" {
			if i+1 >= len(args) {
				return nil, false, nil, errors.New("-p requires a value")
			}
			i++
			params = append(params, args[i])
			continue
		}
		// -pKEY=value (require '=' in suffix; avoid gobbling -path, -print, etc.)
		if strings.HasPrefix(a, "-p") && len(a) > 2 && a[2] != '=' && strings.Contains(a[2:], "=") {
			params = append(params, a[2:])
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered, wantHelp, params, nil
}

// reorderArgs moves flag-like arguments before positional arguments so Go's
// standard flag package can parse them regardless of where the user placed them.
// Flags using --flag=value syntax are handled as a single token. Flags using
// --flag value syntax consume the next argument as the value (unless it also
// starts with "-"). Stops reordering at a bare "--" separator.
func reorderArgs(args []string) []string {
	var flags, positionals []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if !strings.Contains(args[i], "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positionals = append(positionals, args[i])
		}
	}
	return append(flags, positionals...)
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
  -p, --param (preprocessed, not a flag)
        Fraglet-meta parameters as KEY=value (repeatable; any position before "--"). Forms include
        -p K=V, --param K=V, -p=K=V, --param=K=V, and -pK=V when '=' appears in the suffix.
        Optional encodings: -p key=b64:...  See --fraglet-help on a script for its declarations.
  --fraglet-path string
        Path where code is mounted in container (default: /FRAGLET; long form only)
  -e string
        Environment variable to forward into container (repeatable)
        Use -e FOO to forward host value, -e FOO=bar for explicit value
  --fraglet-help
        Show parameter declarations from fraglet-meta and exit (may appear before or after script-file).
        Like -p/--param, removed from argv before your program runs (any position before "--").
        After "--", --fraglet-help and -p/--param pass through unchanged.
  -m, --mode string
        Fraglet mode (sets FRAGLET_CONFIG=/fraglet-{mode}.yml)

Positional:
  script-file   Path to code file (required if -c not set)
  script-args   Tail arguments for your program inside the container

First, -p/--param/--fraglet-help are removed from argv anywhere before a bare "--". Then normal
flags (-v, -c, …) are parsed and must come before script-file. Example: ./tool.py -p city=paris
--profile prod strips -p; --profile and prod are program argv. Use "--" so -p/--param/--fraglet-help
are not stripped.

Stdin:
  Stdin is always forwarded to the program inside the container.
  Use pipes to send data: echo "hello" | fragletc --vein=python script.py

Subcommands:
  mcp           Start the MCP (Model Context Protocol) server over stdio
                Use with Claude Desktop, Cursor, or any MCP-compatible client
  refresh       Refresh (pull) container images for veins
                Use "fragletc refresh --help" for details
  guide         Show fraglet guide for a vein
                Use "fragletc guide --help" for details
  version       Show build version, commit, and lineage info

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

  # Parameters (short -p)
  fragletc --vein=python -p city=london script.py

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

  # Start MCP server (for AI tool integration)
  fragletc mcp
`)
}
