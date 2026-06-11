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
	"github.com/ofthemachine/fraglet/pkg/engine"
	"github.com/ofthemachine/fraglet/pkg/essence"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/guide"
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

// guideEssenceOpts holds parsed arguments for guide and essence subcommands.
// VeinName and Image are mutually exclusive at validation time (see pkg/guide Run).
type guideEssenceOpts struct {
	VeinName string
	Image    string
	Mode     string
}

var errGuideEssenceUsage = errors.New("show usage")

func parseGuideEssenceArgs(args []string) (guideEssenceOpts, error) {
	var o guideEssenceOpts
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if strings.HasPrefix(a, "-") {
			switch {
			case a == "-h" || a == "--help":
				return guideEssenceOpts{}, errGuideEssenceUsage

			case a == "-i" || a == "--image":
				if o.Image != "" {
					return guideEssenceOpts{}, fmt.Errorf("duplicate --image")
				}
				i++
				if i >= len(args) {
					return guideEssenceOpts{}, fmt.Errorf("-i/--image requires a value")
				}
				o.Image = args[i]
				if o.Image == "" {
					return guideEssenceOpts{}, fmt.Errorf("-i/--image requires a non-empty value")
				}

			case strings.HasPrefix(a, "--image="):
				if o.Image != "" {
					return guideEssenceOpts{}, fmt.Errorf("duplicate --image")
				}
				o.Image = strings.TrimPrefix(a, "--image=")
				if o.Image == "" {
					return guideEssenceOpts{}, fmt.Errorf("--image requires a non-empty value")
				}

			case a == "-m" || a == "--mode":
				if o.Mode != "" {
					return guideEssenceOpts{}, fmt.Errorf("duplicate --mode")
				}
				i++
				if i >= len(args) {
					return guideEssenceOpts{}, fmt.Errorf("-m/--mode requires a value")
				}
				o.Mode = args[i]

			case strings.HasPrefix(a, "--mode="):
				if o.Mode != "" {
					return guideEssenceOpts{}, fmt.Errorf("duplicate --mode")
				}
				o.Mode = strings.TrimPrefix(a, "--mode=")

			default:
				return guideEssenceOpts{}, fmt.Errorf("unknown flag %q", a)
			}
			continue
		}
		pos = append(pos, a)
	}

	switch len(pos) {
	case 0:
	case 1:
		o.VeinName = pos[0]
	default:
		return guideEssenceOpts{}, fmt.Errorf("unexpected arguments after vein name: %q", strings.Join(pos[1:], " "))
	}

	return o, nil
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
		case "essence":
			handleEssence()
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
	mode := flag.String("mode", "", "Fraglet mode (sets FRAGLET_MODE=mode)")
	inlineCode := flag.String("c", "", "Program passed in as string (like python -c)")
	var envFlags envListFlag
	flag.Var(&envFlags, "e", "Environment variable to forward (repeatable, e.g. -e FOO -e BAR=val)")

	// Short forms
	flag.StringVar(veinSpec, "v", "", "Vein name with optional mode (short form)")
	flag.StringVar(image, "i", "", "Container image (short form)")
	flag.StringVar(mode, "m", "", "Fraglet mode (short form)")
	flag.StringVar(inlineCode, "code", "", "Program passed in as string (like python -c)")

	// Preprocess argv for params and help
	filtered, wantFragletHelp, paramStrs, err := preprocessFragletArgv(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	if err := flag.CommandLine.Parse(filtered); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	args := flag.Args()
	var scriptFile string
	var scriptArgs []string
	if len(args) > 0 {
		scriptFile = args[0]
		scriptArgs = args[1:]
	}

	if wantFragletHelp {
		handleFragletHelp(scriptFile, *inlineCode)
		return
	}

	var stdinReader io.Reader
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		stdinReader = os.Stdin
	}

	opts := engine.RunOptions{
		VeinSpec:    *veinSpec,
		Image:       *image,
		FragletPath: *fragletPath,
		Mode:        *mode,
		InlineCode:  *inlineCode,
		EnvFlags:    envFlags,
		ScriptFile:  scriptFile,
		ScriptArgs:  scriptArgs,
		Stdin:       stdinReader,
		ParamStrs:   paramStrs,
	}

	exitCode, err := engine.Run(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

const fragletHelpArg = "--fraglet-help"

// argParser implements a simple stateful parser for argument preprocessing
type argParser struct {
	args        []string
	pos         int
	passthrough bool
	filtered    []string
	params      []string
	wantHelp    bool
}

func (p *argParser) peek() (string, bool) {
	if p.pos >= len(p.args) {
		return "", false
	}
	return p.args[p.pos], true
}

func (p *argParser) consume() (string, bool) {
	arg, ok := p.peek()
	if ok {
		p.pos++
	}
	return arg, ok
}

func preprocessFragletArgv(args []string) ([]string, bool, []string, error) {
	p := &argParser{args: args}

	for {
		arg, ok := p.consume()
		if !ok {
			break
		}

		if p.passthrough {
			p.filtered = append(p.filtered, arg)
			continue
		}

		switch {
		case arg == "--":
			p.passthrough = true
			p.filtered = append(p.filtered, arg)

		case arg == fragletHelpArg:
			p.wantHelp = true

		case strings.HasPrefix(arg, "--param="):
			p.params = append(p.params, arg[len("--param="):])

		case arg == "--param":
			val, ok := p.consume()
			if !ok {
				return nil, false, nil, errors.New("--param requires a value")
			}
			p.params = append(p.params, val)

		case strings.HasPrefix(arg, "-p="):
			p.params = append(p.params, arg[len("-p="):])

		case arg == "-p":
			val, ok := p.consume()
			if !ok {
				return nil, false, nil, errors.New("-p requires a value")
			}
			p.params = append(p.params, val)

		case strings.HasPrefix(arg, "-p") && len(arg) > 2 && arg[2] != '=' && strings.Contains(arg[2:], "="):
			// -pKEY=value
			p.params = append(p.params, arg[2:])

		default:
			p.filtered = append(p.filtered, arg)
		}
	}

	return p.filtered, p.wantHelp, p.params, nil
}

func handleFragletHelp(scriptFile, inlineCode string) {
	code := inlineCode
	if code == "" && scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", scriptFile, err)
			os.Exit(1)
		}
		code = string(data)
	}
	if code == "" {
		fmt.Fprintf(os.Stderr, "Error: --fraglet-help requires a script file or -c code\n")
		os.Exit(1)
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

func fragletHelpLabel(scriptFile string) string {
	if scriptFile == "" {
		return "<inline>"
	}
	return filepath.Base(scriptFile)
}

func envVarArrow(d fraglet.ParamDecl) string {
	defaultEnv := strings.ToUpper(d.Alias)
	if d.EnvVar != defaultEnv {
		return fmt.Sprintf("    → %s", d.EnvVar)
	}
	return ""
}

func printFragletInvokeHint(label string) {
	if label == "<inline>" {
		fmt.Fprintf(os.Stdout, "\nPass: fragletc --vein=<vein> -p name=value ... -c '<code>'\n")
		return
	}
	fmt.Fprintf(os.Stdout, "\nPass: ./%s -p name=value ...  (repeat -p per parameter; see fragletc --help)\n", label)
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

	registry, err := vein.LoadAuto(embed.LoadEmbeddedVeins)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
		os.Exit(1)
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
				fmt.Fprintf(os.Stderr, "Error: vein not found: %s\n", name)
				os.Exit(1)
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
		// We use exec here directly as it's a CLI-only helper
		// #nosec G204
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

func printGuideUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  fragletc guide [options and vein/image in any order]

  Resolve image from vein registry: provide exactly one vein name as a positional argument.
  Use image directly: pass -i or --image (no vein name).

Show the fraglet guide from the container's configured guide path.

Options:
  -i, --image string   Container image (mutually exclusive with vein name positional)
  -m, --mode string    Fraglet mode (sets FRAGLET_MODE=mode)
  -h, --help           Show this message

Examples:
  fragletc guide ada
  fragletc guide ada --mode main
  fragletc guide --mode main ada
  fragletc guide -i my-registry/py:latest

The command respects FRAGLET_VEINS_PATH when resolving the vein name.
`)
}

func handleGuide() {
	opts, err := parseGuideEssenceArgs(os.Args[2:])
	if errors.Is(err, errGuideEssenceUsage) {
		printGuideUsage()
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	if opts.Image == "" && opts.VeinName == "" {
		printGuideUsage()
		os.Exit(1)
	}

	var registry *vein.VeinRegistry
	if opts.VeinName != "" {
		registry, err = vein.LoadAuto(embed.LoadEmbeddedVeins)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
			os.Exit(1)
		}
	}

	result, err := guide.Run(context.Background(), registry, opts.VeinName, opts.Mode, opts.Image)
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

func printEssenceUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  fragletc essence [options and vein/image in any order]

  Resolve image from vein registry: provide exactly one vein name as a positional argument.
  Use image directly: pass -i or --image (no vein name).

Show the fraglet essence (short capability summary) from the container.

Options:
  -i, --image string   Container image (mutually exclusive with vein name positional)
  -m, --mode string    Fraglet mode (sets FRAGLET_MODE=mode)
  -h, --help           Show this message

Examples:
  fragletc essence ada
  fragletc essence ada --mode main
  fragletc essence --mode main ada
  fragletc essence -i my-registry/py:latest

The command respects FRAGLET_VEINS_PATH when resolving the vein name.
`)
}

func handleEssence() {
	opts, err := parseGuideEssenceArgs(os.Args[2:])
	if errors.Is(err, errGuideEssenceUsage) {
		printEssenceUsage()
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	if opts.Image == "" && opts.VeinName == "" {
		printEssenceUsage()
		os.Exit(1)
	}

	var registry *vein.VeinRegistry
	if opts.VeinName != "" {
		registry, err = vein.LoadAuto(embed.LoadEmbeddedVeins)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading veins: %v\n", err)
			os.Exit(1)
		}
	}

	result, err := essence.Run(context.Background(), registry, opts.VeinName, opts.Mode, opts.Image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running essence: %v\n", err)
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
        Fraglet mode (sets FRAGLET_MODE=mode)

Positional:
  script-file   Path to code file (required if -c not set)
  script-args   Tail arguments for your program inside the container

First, -p/--param/--fraglet-help are removed from argv anywhere before a bare "--". Then normal
flags (-v, -c, …) are parsed and must come before script-file. Example: ./tool.py -p city=paris
--profile prod strips -p; --profile and prod are program argv. Use "--" so -p/--param/--fraglet-help
are not stripped.

Stdin:
  Stdin is always forwarded to the program inside the container.
  Cat data.csv | ./process.py --format=json

Subcommands:
  mcp           Start the MCP (Model Context Protocol) server over stdio
                Use with Claude Desktop, Cursor, or any MCP-compatible client
  refresh       Refresh (pull) container images for veins
                Use "fragletc refresh --help" for details
  guide         Show fraglet guide (vein registry or --image; flags and vein in any order)
                Use "fragletc guide --help" for details
  essence       Show fraglet essence (vein registry or --image; flags and vein in any order)
                Use "fragletc essence --help" for details
  version       Show build version, commit, and lineage info
`)
}
