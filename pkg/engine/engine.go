package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

const defaultFragletPath = "/FRAGLET"

// RunOptions defines the parameters for executing a fraglet
type RunOptions struct {
	VeinSpec    string
	Image       string
	FragletPath string
	Mode        string
	InlineCode  string
	EnvFlags    []string
	ScriptFile  string
	ScriptArgs  []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	ParamStrs   []string
}

// Run orchestrates the execution of a fraglet
func Run(ctx context.Context, opts RunOptions) (int, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.FragletPath == "" {
		opts.FragletPath = defaultFragletPath
	}

	// --- Resolve vein + mode ---
	veinName, finalMode, err := resolveVeinAndMode(opts.VeinSpec, opts.Mode, opts.Image, opts.ScriptFile)
	if err != nil {
		return 1, fmt.Errorf("Error: %w", err)
	}

	// Validate mutual exclusion early
	if opts.Image != "" && veinName != "" {
		return 1, fmt.Errorf("Error: cannot specify both --image and --vein")
	}

	// --- Resolve code ---
	code, err := resolveCode(opts.InlineCode, opts.ScriptFile)
	if err != nil {
		return 1, fmt.Errorf("Error: %w", err)
	}

	// --- Resolve container + fraglet mount path ---
	containerImage, fragletMountPath, err := resolveContainer(veinName, opts.Image, opts.FragletPath)
	if err != nil {
		return 1, fmt.Errorf("Error: %w", err)
	}

	// --- Build env vars ---
	envVars := buildEnvVars(finalMode, opts.EnvFlags)

	// --- Parse and resolve params ---
	if len(opts.ParamStrs) > 0 {
		var params fraglet.Params
		for _, pf := range opts.ParamStrs {
			p, err := fraglet.ParseParam(pf)
			if err != nil {
				return 1, fmt.Errorf("param error: %w", err)
			}
			params = append(params, p)
		}
		// Resolve aliases via fraglet-meta declarations if code is available
		decls := fraglet.ParseParamDecls(code)
		if len(decls) > 0 {
			var err error
			params, err = params.ResolveAliases(decls)
			if err != nil {
				return 1, fmt.Errorf("param alias error: %w", err)
			}
		}
		transportEnv, err := params.ToTransportEnv()
		if err != nil {
			return 1, fmt.Errorf("param transport error: %w", err)
		}
		envVars = append(envVars, transportEnv...)
	}

	// --- Write temp file, build spec, execute ---
	tmpFile, cleanup, err := writeTempFile(code)
	if err != nil {
		return 1, fmt.Errorf("error creating temp file: %w", err)
	}
	defer cleanup()

	r := runner.NewRunner(containerImage, "")
	spec := runner.RunSpec{
		Container:   containerImage,
		Env:         envVars,
		Args:        opts.ScriptArgs,
		StdinReader: opts.Stdin,
		Stdout:      opts.Stdout,
		Stderr:      opts.Stderr,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: fragletMountPath,
			},
		},
	}

	result, err := r.Run(ctx, spec)
	if err != nil {
		return 1, fmt.Errorf("execution failed: %w", err)
	}

	return result.ExitCode, nil
}

func resolveVeinAndMode(veinSpec, modeFlag, image, scriptFile string) (veinName, mode string, err error) {
	if veinSpec != "" {
		var parsedMode string
		veinName, parsedMode, err = parseVeinSpec(veinSpec)
		if err != nil {
			return "", "", err
		}
		if modeFlag != "" && parsedMode != "" {
			return "", "", fmt.Errorf("mode specified both in --vein and --mode flags")
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
			return "", "", fmt.Errorf("error loading veins: %w", err)
		}
		extMap := vein.NewExtensionMap(registry)
		veinName, err = extMap.VeinForFile(scriptFile)
		if err != nil {
			return "", "", fmt.Errorf("error: %w", err)
		}
	}

	return
}

func resolveCode(inlineCode, scriptFile string) (string, error) {
	if inlineCode != "" {
		return inlineCode, nil
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", scriptFile, err)
		}
		return stripShebang(string(data)), nil
	}
	return "", fmt.Errorf("no code source provided. Use a script file or -c flag")
}

func resolveContainer(veinName, image, fragletPath string) (containerImage, mountPath string, err error) {
	if veinName != "" {
		registry, err := loadVeinRegistry()
		if err != nil {
			return "", "", fmt.Errorf("error loading veins: %w", err)
		}
		v, ok := registry.Get(veinName)
		if !ok {
			return "", "", fmt.Errorf("vein not found: %s", veinName)
		}
		return v.ContainerImage(), defaultFragletPath, nil
	}

	if image != "" {
		return image, fragletPath, nil
	}

	return "", "", fmt.Errorf("no container target. Specify --vein or --image")
}

func buildEnvVars(mode string, envFlags []string) []string {
	var envVars []string
	if mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_MODE=%s", mode))
	}
	for _, entry := range envFlags {
		if strings.Contains(entry, "=") {
			envVars = append(envVars, entry)
		} else {
			if val, ok := os.LookupEnv(entry); ok {
				envVars = append(envVars, entry+"="+val)
			}
		}
	}
	return envVars
}

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
	_ = os.Chmod(tmpFile.Name(), 0644)

	absPath, _ := filepath.Abs(tmpFile.Name())
	cleanup := func() { _ = os.Remove(absPath) }
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
