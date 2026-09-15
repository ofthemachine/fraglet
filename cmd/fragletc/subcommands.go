package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/essence"
	"github.com/ofthemachine/fraglet/pkg/guide"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

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
