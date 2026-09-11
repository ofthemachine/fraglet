package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

const defaultFragletPath = "/FRAGLET"

// OutputMount is the container path where a fraglet writes declared output
// files (fraglet-meta output= decls); ExecuteSpec.OutputHostDir, when set,
// is mounted writable here.
const OutputMount = "/output"

// InputMount is the container path root where file-shaped params
// (fraglet-meta param=<alias>:file) are mounted read-only, one per alias:
// InputMount + "/" + alias.
const InputMount = "/input"

// ExecuteSpec is the explicit, library-shaped input to Execute: no vein
// resolution and no CLI argv parsing, just what to run and how.
type ExecuteSpec struct {
	Image string
	Code  string // full file content (header + body); Execute mounts only the body (see fraglet.SplitHeader)

	Env     []string
	Volumes []runner.VolumeMount // extra mounts: file-shaped params, bundled content, etc.

	// OutputHostDir, when non-empty, is mounted writable at OutputMount.
	// Execute does not create or clean up this directory — that's the
	// caller's job (its lifecycle is tied to what the caller does with the
	// files afterward: capture into CAS, copy out named files, etc.).
	OutputHostDir string

	// SecretEnvNames lists Env entries to redact from verbose logging. Real
	// values still flow to the container unredacted — this only affects
	// what Execute prints when Verbose is set.
	SecretEnvNames []string

	NetworkMode string
	Args        []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Verbose     bool

	// FragletPath is the container mount path for the body script.
	// Defaults to defaultFragletPath when empty.
	FragletPath string
}

// ExecuteResult is the structured outcome of a container run.
type ExecuteResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Execute mounts the body of spec.Code (spec.Code's header — shebang plus
// fraglet-meta/x-operon/plain "#" lines — is stripped via fraglet.SplitHeader
// and never reaches the container) and runs it in spec.Image. It captures
// stdout/stderr into the returned ExecuteResult while also forwarding to
// spec.Stdout/spec.Stderr when set, so a caller gets both a full string
// (e.g. for ledger recording) and live streaming from the same run.
func Execute(ctx context.Context, spec ExecuteSpec) (ExecuteResult, error) {
	networkMode := resolveNetworkMode(spec.NetworkMode, spec.Code)
	return runContainer(ctx, containerRunSpec{
		image:          spec.Image,
		code:           spec.Code,
		env:            spec.Env,
		volumes:        spec.Volumes,
		outputHostDir:  spec.OutputHostDir,
		networkMode:    networkMode,
		args:           spec.Args,
		stdin:          spec.Stdin,
		stdout:         spec.Stdout,
		stderr:         spec.Stderr,
		verbose:        spec.Verbose,
		secretEnvNames: spec.SecretEnvNames,
		fragletPath:    spec.FragletPath,
		capture:        true,
	})
}

// containerRunSpec is the internal, shared execution input for Execute (capture)
// and Run (stream-only).
type containerRunSpec struct {
	image          string
	code           string
	env            []string
	volumes        []runner.VolumeMount
	outputHostDir  string
	networkMode    string
	args           []string
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	verbose        bool
	secretEnvNames []string
	fragletPath    string
	capture        bool
}

func runContainer(ctx context.Context, spec containerRunSpec) (ExecuteResult, error) {
	fragletPath := spec.fragletPath
	if fragletPath == "" {
		fragletPath = defaultFragletPath
	}

	_, body := fraglet.SplitHeader(spec.code)

	if spec.verbose {
		fmt.Fprintf(os.Stderr, "fraglet: executing %s\n", spec.image)
		fmt.Fprintf(os.Stderr, "fraglet: env: %v\n", redactEnv(spec.env, spec.secretEnvNames))
	}

	tmpFile, cleanup, err := writeTempFile(body)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("error creating temp file: %w", err)
	}
	defer cleanup()

	volumes := make([]runner.VolumeMount, 0, len(spec.volumes)+2)
	volumes = append(volumes, runner.VolumeMount{HostPath: tmpFile, ContainerPath: fragletPath})
	volumes = append(volumes, spec.volumes...)
	if spec.outputHostDir != "" {
		volumes = append(volumes, runner.VolumeMount{HostPath: spec.outputHostDir, ContainerPath: OutputMount, Writable: true})
	}

	sink := newOutputSink(spec.capture, spec.stdout, spec.stderr)

	r := runner.NewRunner(spec.image, "")
	result, err := r.Run(ctx, runner.RunSpec{
		Container:   spec.image,
		Env:         spec.env,
		Args:        spec.args,
		NetworkMode: spec.networkMode,
		StdinReader: spec.stdin,
		Stdout:      sink.stdout,
		Stderr:      sink.stderr,
		Volumes:     volumes,
	})
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("execution failed: %w", err)
	}

	out := ExecuteResult{
		ExitCode: result.ExitCode,
		Duration: result.Duration,
	}
	if spec.capture {
		out.Stdout = bufferToString(sink.stdoutBuf, result.Stdout)
		out.Stderr = bufferToString(sink.stderrBuf, result.Stderr)
	}
	return out, nil
}

type outputSink struct {
	stdout    io.Writer
	stderr    io.Writer
	stdoutBuf *bytes.Buffer
	stderrBuf *bytes.Buffer
}

func newOutputSink(capture bool, stdout, stderr io.Writer) outputSink {
	if !capture {
		s := outputSink{stdout: stdout, stderr: stderr}
		if s.stdout == nil {
			s.stdout = io.Discard
		}
		if s.stderr == nil {
			s.stderr = io.Discard
		}
		return s
	}

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	stdoutW := io.Writer(stdoutBuf)
	if stdout != nil {
		stdoutW = io.MultiWriter(stdoutBuf, stdout)
	}
	stderrW := io.Writer(stderrBuf)
	if stderr != nil {
		stderrW = io.MultiWriter(stderrBuf, stderr)
	}
	return outputSink{
		stdout:    stdoutW,
		stderr:    stderrW,
		stdoutBuf: stdoutBuf,
		stderrBuf: stderrBuf,
	}
}

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
	NetworkMode string // docker --network value (e.g. "none" to disable networking); empty = default

	// OutputHostDir, when non-empty, is mounted writable at OutputMount so
	// the fraglet's declared output= files land there. Creating and
	// cleaning up this directory is the caller's job (see cmd/fragletc's
	// --output flag for the CLI-facing named-copy convention).
	OutputHostDir string
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

	// --- Parse and resolve params (including file-shaped params) ---
	var volumes []runner.VolumeMount
	decls := fraglet.ParseParamDecls(code)
	var params fraglet.Params
	for _, pf := range opts.ParamStrs {
		p, err := fraglet.ParseParam(pf)
		if err != nil {
			return 1, fmt.Errorf("param error: %w", err)
		}
		params = append(params, p)
	}
	params = fraglet.ApplyDefaults(decls, params)
	if len(params) > 0 {
		if len(decls) > 0 {
			params, err = params.ResolveAliases(decls)
			if err != nil {
				return 1, fmt.Errorf("param alias error: %w", err)
			}
		}
		var fileMounts []fraglet.FileMount
		params, fileMounts, err = fraglet.ResolveFileParams(decls, params, InputMount)
		if err != nil {
			return 1, fmt.Errorf("file param error: %w", err)
		}
		for _, m := range fileMounts {
			volumes = append(volumes, runner.VolumeMount{HostPath: m.HostPath, ContainerPath: m.ContainerPath})
		}
		transportEnv, err := params.ToTransportEnv()
		if err != nil {
			return 1, fmt.Errorf("param transport error: %w", err)
		}
		envVars = append(envVars, transportEnv...)
	}

	// --- Resolve network mode ---
	networkMode := resolveNetworkMode(opts.NetworkMode, code)

	result, err := runContainer(ctx, containerRunSpec{
		image:         containerImage,
		code:          code,
		env:           envVars,
		volumes:       volumes,
		outputHostDir: opts.OutputHostDir,
		networkMode:   networkMode,
		args:          opts.ScriptArgs,
		stdin:         opts.Stdin,
		stdout:        opts.Stdout,
		stderr:        opts.Stderr,
		fragletPath:   fragletMountPath,
		capture:       false,
	})
	if err != nil {
		return 1, err
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

// resolveCode returns the full file content unmodified. Execute (via
// fraglet.SplitHeader) strips the header — shebang plus all fraglet-meta /
// x-operon / plain "#" lines — before the container ever sees any of it, so
// there is nothing left for the caller to strip here.
func resolveCode(inlineCode, scriptFile string) (string, error) {
	if inlineCode != "" {
		return inlineCode, nil
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", scriptFile, err)
		}
		return string(data), nil
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

// resolveNetworkMode determines the effective container network mode.
// An explicit mode (from CLI flag or ExecuteSpec) takes precedence.
// If unspecified, a script declaring "#: network=none" defaults to "none".
// Otherwise, returns "" (which directs the runner to omit --network, leaving
// Docker to its default bridge).
func resolveNetworkMode(explicitMode, code string) string {
	if explicitMode != "" {
		return explicitMode
	}
	if fraglet.ParseNetwork(code) == "none" {
		return "none"
	}
	return ""
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

func bufferToString(buf *bytes.Buffer, fallback string) string {
	if buf.Len() > 0 {
		return buf.String()
	}
	return fallback
}

// redactEnv returns env with the value of every entry whose name is in
// secrets replaced by "***". Used only for verbose logging — the real
// values are always what reaches the container.
func redactEnv(env []string, secrets []string) []string {
	if len(secrets) == 0 {
		return env
	}
	secretSet := make(map[string]bool, len(secrets))
	for _, s := range secrets {
		secretSet[s] = true
	}
	redacted := make([]string, len(env))
	for i, e := range env {
		name, _, ok := strings.Cut(e, "=")
		if ok && secretSet[name] {
			redacted[i] = name + "=***"
		} else {
			redacted[i] = e
		}
	}
	return redacted
}
