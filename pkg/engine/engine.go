package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ofthemachine/fraglet/pkg/embed"
	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/receipt"
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

	// Secrets are the values of the script's "#: secret=" declarations, by
	// name: the caller (a host, which knows where its values live) supplies
	// them; Execute never reads its own process environment for them. Each
	// declared one is delivered as fragletc delivers it, the read-only file
	// fraglet.SecretMountDir/NAME with its path in NAME_FILE, never as an env
	// var; a declared name missing here, or one Env also sets, is refused.
	// Values for names the script does not declare are not delivered.
	Secrets map[string]string

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
	// SecretOutput is what the run's secrets did in what it wrote (its
	// script's secret-output= mode); nil when it was given none.
	SecretOutput *receipt.SecretOutput
}

// Execute mounts the body of spec.Code (spec.Code's header — shebang plus
// "#:" directive and plain "#" lines — is stripped via fraglet.SplitHeader
// and never reaches the container) and runs it in spec.Image. It captures
// stdout/stderr into the returned ExecuteResult while also forwarding to
// spec.Stdout/spec.Stderr when set, so a caller gets both a full string
// (e.g. to record the run) and live streaming from the same run.
func Execute(ctx context.Context, spec ExecuteSpec) (ExecuteResult, error) {
	networkMode := resolveNetworkMode(spec.NetworkMode, spec.Code)
	secrets, secretEnv, err := SecretsFor(spec.Code, spec.Env, spec.Secrets)
	if err != nil {
		return ExecuteResult{}, err
	}
	res, err := runContainer(ctx, containerRunSpec{
		image:          spec.Image,
		code:           spec.Code,
		env:            append(slices.Clone(spec.Env), secretEnv...),
		secrets:        secrets,
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
	if err == nil {
		err = secretOutputErr(res.SecretOutput)
	}
	return res, err
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
	secrets        []runner.Secret
}

func runContainer(ctx context.Context, spec containerRunSpec) (ExecuteResult, error) {
	fragletPath := spec.fragletPath
	if fragletPath == "" {
		fragletPath = defaultFragletPath
	}

	_, body := fraglet.SplitHeader(spec.code)
	mode, err := secretOutputMode(spec.code)
	if err != nil {
		return ExecuteResult{}, err
	}

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

	// What the run writes passes its secrets' guard before anything sees
	// it (the caller, a capture, a hash): its script answers for what it
	// prints (#: secret-output=).
	sink := newOutputSink(spec.capture, spec.stdout, spec.stderr)
	stdout := newSecretGuard(sink.stdout, "stdout", spec.secrets, mode)
	stderr := newSecretGuard(sink.stderr, "stderr", spec.secrets, mode)

	r := runner.NewRunner(spec.image, "")
	result, err := r.Run(ctx, runner.RunSpec{
		Container:   spec.image,
		Env:         spec.env,
		Args:        spec.args,
		NetworkMode: spec.networkMode,
		StdinReader: spec.stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		Volumes:     volumes,
		Secrets:     spec.secrets,
	})
	if cerr := errors.Join(stdout.close(), stderr.close()); err == nil {
		err = cerr
	}
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("execution failed: %w", err)
	}
	filed, err := guardOutputs(spec.outputHostDir, spec.secrets, mode)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("outputs: %w", err)
	}

	out := ExecuteResult{
		ExitCode:     result.ExitCode,
		Duration:     result.Duration,
		SecretOutput: secretOutput(mode, spec.secrets, slices.Concat(stdout.seen, stderr.seen, filed)),
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
	// StdinMode overrides the script's #: stdin= declaration (fraglet.StdinNone,
	// StdinBuffer, StdinStream); empty defers to the header, and an undeclared
	// header means StdinBuffer: Stdin is read to EOF before the run, hashed
	// into the report, and forwarded. Stream passes it through live and
	// hashes nothing; None attaches nothing.
	StdinMode string

	// OutputHostDir, when non-empty, is mounted writable at OutputMount so
	// the fraglet's declared output= files land there. Creating and
	// cleaning up this directory is the caller's job (see cmd/fragletc's
	// --output flag for the CLI-facing named-copy convention).
	OutputHostDir string

	// StreamsDir, when non-empty, receives the run's stdout and stderr as
	// the files "stdout" and "stderr": exactly the bytes the report hashes
	// (after the secret-output mode), the caller's own messages never among
	// them, so whoever publishes a receipt can publish what it names.
	StreamsDir string

	// Now is the instant dynamic defaults (default=@hour) resolve against;
	// zero means the clock. A caller running a chain passes one Now to every
	// run, so its steps share one bucket.
	Now time.Time
}

// Run orchestrates the execution of a fraglet
func Run(ctx context.Context, opts RunOptions) (int, error) {
	report, err := RunWithReport(ctx, opts)
	return report.ExitCode, err
}

// RunReport is everything RunWithReport learned about a run: the
// Invocation that was bound to the container and the Outcome it produced,
// in receipt's vocabulary. receipt.Build turns it into a receipt.
type RunReport struct {
	receipt.Invocation
	receipt.Outcome
}

// RunWithReport is Run plus a RunReport. Stdout is hashed as it streams;
// stdin is bound per the effective stdin mode (see RunOptions.StdinMode).
// On error ExitCode is 1 unless the container itself reported otherwise.
func RunWithReport(ctx context.Context, opts RunOptions) (RunReport, error) {
	report, err := runWithReport(ctx, opts)
	if err != nil && report.ExitCode == 0 {
		report.ExitCode = 1
	}
	return report, err
}

func runWithReport(ctx context.Context, opts RunOptions) (RunReport, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.FragletPath == "" {
		opts.FragletPath = defaultFragletPath
	}

	spec, inv, anonStdout, err := prepare(opts)
	if err != nil {
		return RunReport{Invocation: inv}, err
	}

	stdoutHash := receipt.NewHasher()
	stdoutTo := []io.Writer{opts.Stdout, stdoutHash}
	stderrHash, stderrTail := receipt.NewHasher(), &receipt.Tail{}
	stderrTo := []io.Writer{opts.Stderr, stderrHash, stderrTail}
	var streams []*os.File
	if opts.StreamsDir != "" {
		if streams, err = createStreams(opts.StreamsDir); err != nil {
			return RunReport{Invocation: inv}, fmt.Errorf("streams: %w", err)
		}
		stdoutTo, stderrTo = append(stdoutTo, streams[0]), append(stderrTo, streams[1])
	}
	spec.stdout = io.MultiWriter(stdoutTo...)
	spec.stderr = io.MultiWriter(stderrTo...)

	var out receipt.Outcome
	out.Started = time.Now().UTC()
	result, err := runContainer(ctx, spec)
	out.Finished = time.Now().UTC()
	for _, f := range streams {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = fmt.Errorf("streams: %w", cerr)
		}
	}
	out.Stderr = stderrHash.Digest()
	if err != nil || result.ExitCode != 0 {
		masks := make([][]byte, 0, len(spec.secrets))
		for _, s := range spec.secrets {
			masks = append(masks, s.Value)
		}
		out.StderrTail = stderrTail.String(masks...)
	}
	if err != nil {
		return RunReport{Invocation: inv, Outcome: out}, err
	}
	out.ExitCode = result.ExitCode
	out.Stdout = stdoutHash.Digest()
	out.SecretOutput = result.SecretOutput
	if out.Outputs, err = collectOutputs(opts.OutputHostDir, anonStdout, out.Stdout); err != nil {
		return RunReport{Invocation: inv, Outcome: out}, fmt.Errorf("outputs: %w", err)
	}
	// The image is local by now (the run pulled it if needed), so its
	// registry digest is a cheap inspect away. ResolveImageDigest returns
	// the ref unchanged when it cannot; only a real @sha256 counts.
	if ref, _ := vein.ResolveImageDigest(ctx, spec.image); strings.Contains(ref, "@sha256:") {
		inv.ImageDigest = ref[strings.Index(ref, "@")+1:]
	}
	return RunReport{Invocation: inv, Outcome: out}, secretOutputErr(out.SecretOutput)
}

// prepare resolves everything a run needs into the container spec and, from
// that same final state, the Invocation that identifies it. anonStdout
// reports whether stdout is the run's anonymous result (no output= declared).
func prepare(opts RunOptions) (spec containerRunSpec, inv receipt.Invocation, anonStdout bool, err error) {
	// --- Resolve vein + mode ---
	veinName, finalMode, err := resolveVeinAndMode(opts.VeinSpec, opts.Mode, opts.Image, opts.ScriptFile)
	if err != nil {
		return spec, inv, false, fmt.Errorf("Error: %w", err)
	}

	// Validate mutual exclusion early
	if opts.Image != "" && veinName != "" {
		return spec, inv, false, fmt.Errorf("Error: cannot specify both --image and --vein")
	}

	// --- Resolve code ---
	code, err := resolveCode(opts.InlineCode, opts.ScriptFile)
	if err != nil {
		return spec, inv, false, fmt.Errorf("Error: %w", err)
	}

	// --- Resolve container + fraglet mount path ---
	containerImage, fragletMountPath, err := resolveContainer(veinName, opts.Image, opts.FragletPath)
	if err != nil {
		return spec, inv, false, fmt.Errorf("Error: %w", err)
	}

	// --- Build env vars ---
	envVars := buildEnvVars(finalMode, opts.EnvFlags)

	// --- Resolve declared secrets: values from the caller's env, delivered as files ---
	secrets, secretEnv, err := resolveSecrets(fraglet.ParseSecretDecls(code), opts.EnvFlags, os.LookupEnv)
	if err != nil {
		return spec, inv, false, fmt.Errorf("Error: %w", err)
	}
	envVars = append(envVars, secretEnv...)

	// --- Parse and resolve params (including file-shaped params) ---
	var volumes []runner.VolumeMount
	decls := fraglet.ParseParamDecls(code)
	var params fraglet.Params
	for _, pf := range opts.ParamStrs {
		p, err := fraglet.ParseParam(pf)
		if err != nil {
			return spec, inv, false, fmt.Errorf("param error: %w", err)
		}
		params = append(params, p)
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	params, err = fraglet.ApplyDefaults(decls, params, now)
	if err != nil {
		return spec, inv, false, fmt.Errorf("param error: %w", err)
	}
	var scalars, inputs map[string]string
	if len(params) > 0 {
		if len(decls) > 0 {
			params, err = params.ResolveAliases(decls)
			if err != nil {
				return spec, inv, false, fmt.Errorf("param alias error: %w", err)
			}
		}
		// ResolveFileParams validates each :file host path and rewrites the
		// value to its container path; the receipt hashes the host file, so
		// split from the pre-rewrite params after validation has passed.
		resolved := params
		var fileMounts []fraglet.FileMount
		params, fileMounts, err = fraglet.ResolveFileParams(decls, params, InputMount)
		if err != nil {
			return spec, inv, false, fmt.Errorf("file param error: %w", err)
		}
		scalars, inputs, err = splitParams(decls, resolved)
		if err != nil {
			return spec, inv, false, fmt.Errorf("param error: %w", err)
		}
		for _, m := range fileMounts {
			volumes = append(volumes, runner.VolumeMount{HostPath: m.HostPath, ContainerPath: m.ContainerPath})
		}
		transportEnv, err := params.ToTransportEnv()
		if err != nil {
			return spec, inv, false, fmt.Errorf("param transport error: %w", err)
		}
		envVars = append(envVars, transportEnv...)
	}
	if inputs == nil {
		inputs = map[string]string{}
	}

	// --- Bind stdin ---
	stdinMode := opts.StdinMode
	if stdinMode == "" {
		stdinMode = fraglet.ParseStdin(code)
	}
	if stdinMode == "" {
		stdinMode = fraglet.StdinBuffer
	}
	var stdin io.Reader
	switch stdinMode {
	case fraglet.StdinNone:
		inputs[receipt.AnonKey] = receipt.SumBytes(nil).Hash
	case fraglet.StdinBuffer:
		// A nil Stdin (a terminal, or a library caller with nothing to
		// say) is not attached at all: an empty reader would add -i and
		// make docker wait on nothing.
		var stdinBytes []byte
		if opts.Stdin != nil {
			if stdinBytes, err = io.ReadAll(opts.Stdin); err != nil {
				return spec, inv, false, fmt.Errorf("stdin: %w", err)
			}
			stdin = bytes.NewReader(stdinBytes)
		}
		inputs[receipt.AnonKey] = receipt.SumBytes(stdinBytes).Hash
	case fraglet.StdinStream:
		stdin = opts.Stdin
	default:
		return spec, inv, false, fmt.Errorf("Error: stdin mode %q is not none, buffer, or stream", stdinMode)
	}

	spec = containerRunSpec{
		image:         containerImage,
		code:          code,
		env:           envVars,
		volumes:       volumes,
		outputHostDir: opts.OutputHostDir,
		networkMode:   resolveNetworkMode(opts.NetworkMode, code),
		args:          opts.ScriptArgs,
		stdin:         stdin,
		fragletPath:   fragletMountPath,
		capture:       false,
		secrets:       secrets,
	}
	inv = receipt.Invocation{
		ProcedureHash: receipt.ProcedureHash([]byte(code)),
		Image:         containerImage,
		Network:       effectiveNetwork(opts.NetworkMode, code),
		Mode:          finalMode,
		StdinMode:     stdinMode,
		Params:        scalars,
		Inputs:        inputs,
		Argv:          opts.ScriptArgs,
		Env:           envNames(opts.EnvFlags),
	}
	return spec, inv, len(fraglet.ParseOutputDecls(code)) == 0, nil
}

// effectiveNetwork is the network stance the run had: a --network override
// verbatim, else the declared #: network= value ("" when undeclared).
func effectiveNetwork(override, code string) string {
	if override != "" {
		return override
	}
	return fraglet.ParseNetwork(code)
}

// envNames lists the variables -e forwards, by name only.
func envNames(envFlags []string) []string {
	var names []string
	for _, e := range envFlags {
		if name, _, _ := strings.Cut(e, "="); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// createStreams creates dir/stdout and dir/stderr, empty, for a run to
// write (a retry's run replaces the last one's): the caller closes both.
func createStreams(dir string) ([]*os.File, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var files []*os.File
	for _, name := range []string{"stdout", "stderr"} {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			for _, open := range files {
				open.Close()
			}
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// collectOutputs hashes every file the run left under /output; when the
// script declares no output=, stdout is the anonymous result instead
// (receipt.AnonKey). Stdout is recorded on the Outcome either way.
func collectOutputs(outputHostDir string, anonStdout bool, stdout receipt.Digest) (map[string]string, error) {
	outputs := map[string]string{}
	if anonStdout {
		outputs[receipt.AnonKey] = stdout.Hash
	}
	if outputHostDir == "" {
		return outputs, nil
	}
	err := filepath.WalkDir(outputHostDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(outputHostDir, p)
		if err != nil {
			return err
		}
		dg, err := receipt.SumFile(p)
		if err != nil {
			return err
		}
		outputs[filepath.ToSlash(rel)] = dg.Hash
		return nil
	})
	return outputs, err
}

// splitParams splits resolved params into the receipt's two maps: scalar
// params by declared alias and decoded value, and :file params by alias to
// the content hash of the host file they name.
func splitParams(decls []fraglet.ParamDecl, params fraglet.Params) (map[string]string, map[string]string, error) {
	byEnv := make(map[string]fraglet.ParamDecl, len(decls))
	for _, d := range decls {
		byEnv[d.EnvVar] = d
	}
	scalars := make(map[string]string, len(params))
	inputs := make(map[string]string)
	for _, p := range params {
		value, err := p.Decode()
		if err != nil {
			return nil, nil, err
		}
		name := p.EnvVar
		isFile := false
		if d, ok := byEnv[p.EnvVar]; ok {
			name = d.Alias
			isFile = d.Shape == "file"
		}
		if !isFile {
			scalars[name] = value
			continue
		}
		dg, err := receipt.SumFile(value)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", name, err)
		}
		inputs[name] = dg.Hash
	}
	return scalars, inputs, nil
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
// fraglet.SplitHeader) strips the header — shebang plus all "#:" directive
// and plain "#" lines — before the container ever sees any of it, so
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

// SecretsFor resolves code's "#: secret=" declarations against values (a
// caller's, by name) into what a runner delivers: the files, and the
// NAME_FILE env vars that name them. env is what else the run sets; a
// declared name (or its NAME_FILE) in it is refused, and a declared name
// values lacks is missing. Execute uses it, and so may a caller that runs a
// container itself (a bare image), so secrets reach every container the same
// way.
func SecretsFor(code string, env []string, values map[string]string) ([]runner.Secret, []string, error) {
	return resolveSecrets(fraglet.ParseSecretDecls(code), env, func(name string) (string, bool) {
		v, ok := values[name]
		return v, ok
	})
}

// resolveSecrets reads each declared secret through lookup (fragletc: its
// own environment; a library caller: the values it supplies).
// The container gets the value only as a file (runner.Secret) plus the env
// var NAME_FILE holding its path; the value itself is never an env var, so
// forwarding a declared secret with -e is refused rather than quietly
// exposing it. Nothing here reaches the receipt: the declarations are part
// of the script, which procedure_hash already covers.
func resolveSecrets(decls []fraglet.SecretDecl, envFlags []string, lookup func(string) (string, bool)) ([]runner.Secret, []string, error) {
	if len(decls) == 0 {
		return nil, nil, nil
	}
	forwarded := make(map[string]bool)
	for _, name := range envNames(envFlags) {
		forwarded[name] = true
	}
	var secrets []runner.Secret
	var env, missing []string
	for _, d := range decls {
		if !fraglet.ValidSecretName(d.Name) {
			return nil, nil, fmt.Errorf("secret=%s is not a valid secret name (see fragletc lint)", d.Name)
		}
		if forwarded[d.Name] || forwarded[d.FileEnv()] {
			return nil, nil, fmt.Errorf("-e %s: %s is a declared secret; fragletc delivers it as the file $%s, so do not forward it as an environment variable", d.Name, d.Name, d.FileEnv())
		}
		value, ok := lookup(d.Name)
		if !ok || value == "" {
			missing = append(missing, d.Name)
			continue
		}
		secrets = append(secrets, runner.Secret{Name: d.Name, Path: d.ContainerPath(), Value: []byte(value)})
		env = append(env, d.FileEnv()+"="+d.ContainerPath())
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf("missing secret(s): %s -- set them in the environment fragletc runs in (e.g. `op run -- fragletc ...`)", strings.Join(missing, ", "))
	}
	return secrets, env, nil
}
