package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ofthemachine/fraglet/internal/testutil"
	"github.com/ofthemachine/fraglet/pkg/receipt"
	"github.com/ofthemachine/fraglet/pkg/runner"
)

func dockerAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return exec.Command("docker", "version").Run() == nil
}

func requireDocker(t *testing.T) {
	t.Helper()
	if !dockerAvailable(t) {
		t.Skip("docker not available, skipping")
	}
}

func TestExecute_StripsHeaderBeforeMount(t *testing.T) {
	requireDocker(t)

	code := "#!/usr/bin/env -S fragletc --image alpine:latest\n" +
		"#: d=HEADER_MARKER_SHOULD_NOT_APPEAR\n" +
		"#: param=unused\n" +
		"# plain-comment: ref=test/marker\n" +
		"BODY_MARKER_SHOULD_APPEAR"

	result, err := Execute(context.Background(), ExecuteSpec{
		Image: "alpine:latest",
		Code:  code,
		Args:  []string{"sh", "-c", "cat " + defaultFragletPath},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(result.Stdout, "HEADER_MARKER_SHOULD_NOT_APPEAR") {
		t.Fatalf("header leaked into mounted body: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "BODY_MARKER_SHOULD_APPEAR") {
		t.Fatalf("body missing from mounted script: %q", result.Stdout)
	}
	// "#:" directive lines and plain "#" header comments must never reach the container.
	if strings.Contains(result.Stdout, "plain-comment") || strings.Contains(result.Stdout, "#:") {
		t.Fatalf("directive lines leaked into mounted body: %q", result.Stdout)
	}
}

func TestExecute_OutputMount(t *testing.T) {
	requireDocker(t)

	outDir := t.TempDir()
	// t.TempDir() is 0700. Containers run with --cap-drop=all (pkg/runner/
	// docker.go), which strips CAP_DAC_OVERRIDE, so root inside the
	// container no longer bypasses host permission checks — on a real
	// Linux dockerd (not Docker Desktop's more permissive bind mounts),
	// root can't write into a dir it doesn't own unless it's opened up.
	// Execute() itself never chmods OutputHostDir (it isn't Execute's
	// directory to manage), so whoever creates the scratch dir — here,
	// this test, mirroring what cmd/fragletc/fragletc.go does for its own
	// scratch dir — is responsible for making it writable.
	if err := os.Chmod(outDir, 0o777); err != nil {
		t.Fatal(err)
	}
	code := "#!/usr/bin/env -S fragletc --image alpine:latest\n#: output=result.txt\n\necho placeholder"

	_, err := Execute(context.Background(), ExecuteSpec{
		Image:         "alpine:latest",
		Code:          code,
		OutputHostDir: outDir,
		Args:          []string{"sh", "-c", "echo hello > " + OutputMount + "/result.txt"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "result.txt"))
	if err != nil {
		t.Fatalf("read captured output: %v", err)
	}
	if strings.TrimSpace(string(data)) != "hello" {
		t.Fatalf("output content = %q, want hello", data)
	}
}

func TestExecute_FileParamMount(t *testing.T) {
	requireDocker(t)

	hostFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(hostFile, []byte("mounted content"), 0o644); err != nil {
		t.Fatalf("write host file: %v", err)
	}

	code := "#!/usr/bin/env -S fragletc --image alpine:latest\n#: param=doc:file\n\necho placeholder"

	result, err := Execute(context.Background(), ExecuteSpec{
		Image: "alpine:latest",
		Code:  code,
		Volumes: []runner.VolumeMount{
			{HostPath: hostFile, ContainerPath: InputMount + "/doc"},
		},
		Args: []string{"sh", "-c", "cat " + InputMount + "/doc"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "mounted content" {
		t.Fatalf("stdout = %q, want mounted content", result.Stdout)
	}
}

func TestExecute_VerboseRedactsSecrets(t *testing.T) {
	requireDocker(t)

	code := "#!/usr/bin/env -S fragletc --image alpine:latest\n\necho hi"
	_, err := Execute(context.Background(), ExecuteSpec{
		Image:          "alpine:latest",
		Code:           code,
		Env:            []string{"API_KEY=super-secret", "MODE=demo"},
		SecretEnvNames: []string{"API_KEY"},
		Verbose:        true,
		Args:           []string{"sh", "-c", "echo hi"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// redactEnv itself is covered directly below; this test only asserts
	// Execute runs to completion with Verbose set (the log line goes to
	// os.Stderr, not to a capturable writer, so we don't assert its content
	// here — see TestRedactEnv for the redaction logic itself).
}

func TestRedactEnv(t *testing.T) {
	env := []string{"API_KEY=super-secret", "MODE=demo"}
	got := redactEnv(env, []string{"API_KEY"})
	want := []string{"API_KEY=***", "MODE=demo"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("redactEnv = %v, want %v", got, want)
		}
	}
	// No secrets declared: env passes through unchanged (same slice values).
	got = redactEnv(env, nil)
	for i := range env {
		if got[i] != env[i] {
			t.Fatalf("redactEnv with no secrets = %v, want unchanged %v", got, env)
		}
	}
}

func TestResolveNetworkMode(t *testing.T) {
	tests := []struct {
		name         string
		explicitMode string
		code         string
		want         string
	}{
		{
			name:         "empty defaults to bridge",
			explicitMode: "",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest

				print(1)
			`),
			want: "",
		},
		{
			name:         "header network none defaults to none",
			explicitMode: "",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest
				#: network=none

				echo "pure computation"
			`),
			want: "none",
		},
		{
			name:         "header network required defaults to bridge",
			explicitMode: "",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest
				#: network=required

				curl http://example.com
			`),
			want: "",
		},
		{
			name:         "explicit mode overrides header",
			explicitMode: "bridge",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest
				#: network=none

				echo "overridden"
			`),
			want: "bridge",
		},
		{
			name:         "explicit none without header",
			explicitMode: "none",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest

				echo "no header"
			`),
			want: "none",
		},
		{
			name:         "explicit host overrides network none",
			explicitMode: "host",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc --image alpine:latest
				#: network=none

				echo "host network"
			`),
			want: "host",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveNetworkMode(tc.explicitMode, tc.code); got != tc.want {
				t.Fatalf("resolveNetworkMode(%q, code) = %q, want %q", tc.explicitMode, got, tc.want)
			}
		})
	}
}

func TestNewOutputSink_CaptureBuffersAndForwards(t *testing.T) {
	var forwarded strings.Builder
	sink := newOutputSink(true, &forwarded, nil)
	if sink.stdoutBuf == nil || sink.stderrBuf == nil {
		t.Fatal("capture mode must allocate buffers")
	}
	if _, err := sink.stdout.Write([]byte("hello")); err != nil {
		t.Fatalf("stdout write: %v", err)
	}
	if got := sink.stdoutBuf.String(); got != "hello" {
		t.Fatalf("stdoutBuf = %q, want hello", got)
	}
	if got := forwarded.String(); got != "hello" {
		t.Fatalf("forwarded stdout = %q, want hello", got)
	}
}

func TestNewOutputSink_NoCaptureStreamsWithoutBuffer(t *testing.T) {
	var stdout strings.Builder
	sink := newOutputSink(false, &stdout, nil)
	if sink.stdoutBuf != nil || sink.stderrBuf != nil {
		t.Fatal("non-capture mode must not allocate buffers")
	}
	payload := strings.Repeat("x", 4096)
	if _, err := sink.stdout.Write([]byte(payload)); err != nil {
		t.Fatalf("stdout write: %v", err)
	}
	if got := stdout.String(); got != payload {
		t.Fatalf("forwarded stdout len = %d, want %d", len(got), len(payload))
	}
}

// pythonImage is the image the RunWithReport tests run scripts in. They
// skip unless it is already local: these are unit tests, not pulls.
const pythonImage = "ofthemachine/python3:latest"

func requireLocalImage(t *testing.T, image string) {
	t.Helper()
	requireDocker(t)
	if exec.Command("docker", "image", "inspect", image).Run() != nil {
		t.Skipf("image %s not present locally, skipping", image)
	}
}

func writeScript(t *testing.T, name, code string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(code), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func runReport(t *testing.T, opts RunOptions) (RunReport, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	opts.Image, opts.Stdout, opts.Stderr = pythonImage, &stdout, &stderr
	report, err := RunWithReport(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunWithReport: %v\nstderr: %s", err, stderr.String())
	}
	if report.ExitCode != 0 {
		t.Fatalf("exit %d\nstderr: %s", report.ExitCode, stderr.String())
	}
	return report, stdout.String()
}

const upperScript = "#!/usr/bin/env -S fragletc --image ofthemachine/python3\n#: d=upper\nimport sys\nprint(sys.stdin.read().upper(), end='')\n"

// Undeclared stdin is buffered: forwarded to the program and hashed into
// the invocation, so the run has a memo key.
func TestRunWithReport_UndeclaredStdinIsBuffered(t *testing.T) {
	requireLocalImage(t, pythonImage)
	script := writeScript(t, "upper.py", upperScript)

	report, out := runReport(t, RunOptions{ScriptFile: script, Stdin: strings.NewReader("hello\n")})
	if out != "HELLO\n" {
		t.Fatalf("stdout = %q", out)
	}
	if report.StdinMode != "buffer" || report.Inputs[receipt.AnonKey] != receipt.SumBytes([]byte("hello\n")).Hash {
		t.Fatalf("invocation = %+v", report.Invocation)
	}
	if _, ok := report.MemoKey(); !ok {
		t.Fatalf("expected a memo key: unbound = %v", report.Unbound())
	}
	if report.Outputs[receipt.AnonKey] != report.Stdout.Hash || report.Stdout.Bytes != 6 {
		t.Fatalf("outcome = %+v", report.Outcome)
	}
}

// A nil Stdin under buffer is not attached: the program sees EOF and the
// input hash is the empty one, so the key is the same as for no stdin.
func TestRunWithReport_NilStdinNotAttached(t *testing.T) {
	requireLocalImage(t, pythonImage)
	script := writeScript(t, "upper.py", upperScript)

	report, out := runReport(t, RunOptions{ScriptFile: script})
	if out != "" {
		t.Fatalf("stdout = %q", out)
	}
	if report.Inputs[receipt.AnonKey] != receipt.SumBytes(nil).Hash {
		t.Fatalf("inputs = %v", report.Inputs)
	}
}

// Stream and none are explicit: stream forwards live and drops the key;
// none attaches nothing even when a Stdin is offered.
func TestRunWithReport_StreamAndNone(t *testing.T) {
	requireLocalImage(t, pythonImage)
	script := writeScript(t, "upper.py", upperScript)

	report, out := runReport(t, RunOptions{ScriptFile: script, Stdin: strings.NewReader("hi\n"), StdinMode: "stream"})
	if out != "HI\n" || report.StdinMode != "stream" {
		t.Fatalf("stream: stdout = %q, mode = %s", out, report.StdinMode)
	}
	if _, ok := report.Inputs[receipt.AnonKey]; ok {
		t.Fatal("stream must not record a stdin hash")
	}
	if _, ok := report.MemoKey(); ok {
		t.Fatal("stream must not have a memo key")
	}

	report, out = runReport(t, RunOptions{ScriptFile: script, Stdin: strings.NewReader("hi\n"), StdinMode: "none"})
	if out != "" || report.StdinMode != "none" || report.Inputs[receipt.AnonKey] != receipt.SumBytes(nil).Hash {
		t.Fatalf("none: stdout = %q, invocation = %+v", out, report.Invocation)
	}
}

// Anything bound outside the declared parameters is recorded on the
// invocation and costs it its memo key; -e is recorded by name only.
func TestRunWithReport_ArgvAndEnvAreUnbound(t *testing.T) {
	requireLocalImage(t, pythonImage)
	script := writeScript(t, "args.py", "import os, sys\nprint(sys.argv[1:], os.environ.get('SECRET_TOKEN'))\n")

	report, out := runReport(t, RunOptions{ScriptFile: script, ScriptArgs: []string{"foo"}, EnvFlags: []string{"SECRET_TOKEN=hunter2"}})
	if !strings.Contains(out, "['foo'] hunter2") {
		t.Fatalf("stdout = %q", out)
	}
	if len(report.Argv) != 1 || report.Argv[0] != "foo" || len(report.Env) != 1 || report.Env[0] != "SECRET_TOKEN" {
		t.Fatalf("invocation = %+v", report.Invocation)
	}
	if _, ok := report.MemoKey(); ok {
		t.Fatal("argv/env runs must not have a memo key")
	}
	if len(report.Unbound()) != 2 {
		t.Fatalf("Unbound = %v", report.Unbound())
	}
	data, _ := json.Marshal(receipt.Build(report.Invocation, report.Outcome, "test", script))
	if bytes.Contains(data, []byte("hunter2")) {
		t.Fatalf("secret value leaked into receipt: %s", data)
	}
}

// Declared params and :file inputs are the key's whole input space, and a
// conformant run's key is the documented formula over them.
func TestRunWithReport_DeclaredParamsKeyTheRun(t *testing.T) {
	requireLocalImage(t, pythonImage)
	doc := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(doc, []byte("mounted"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := "#!/usr/bin/env -S fragletc --image ofthemachine/python3\n#: param=n:d=count\n#: param=doc:file\n#: output=out.txt\nimport os\nopen('/output/out.txt','w').write(os.environ['N'] + open('/input/doc').read())\n"
	script := writeScript(t, "tool.py", code)
	outDir := t.TempDir()
	if err := os.Chmod(outDir, 0o777); err != nil {
		t.Fatal(err)
	}

	report, _ := runReport(t, RunOptions{ScriptFile: script, ParamStrs: []string{"n=7", "doc=" + doc}, OutputHostDir: outDir})
	if report.Params["n"] != "7" || report.Inputs["doc"] != receipt.SumBytes([]byte("mounted")).Hash {
		t.Fatalf("invocation = %+v", report.Invocation)
	}
	want, _ := receipt.Invocation{ProcedureHash: receipt.ProcedureHash([]byte(code)), Params: map[string]string{"n": "7"}, Inputs: map[string]string{"": receipt.SumBytes(nil).Hash, "doc": receipt.SumBytes([]byte("mounted")).Hash}}.MemoKey()
	if key, ok := report.MemoKey(); !ok || key != want {
		t.Fatalf("memo key = %s (%v), want %s", key, ok, want)
	}
	if report.Outputs["out.txt"] != receipt.SumBytes([]byte("7mounted")).Hash {
		t.Fatalf("outputs = %v", report.Outputs)
	}
	if _, ok := report.Outputs[receipt.AnonKey]; ok {
		t.Fatal("a script with declared outputs has no anonymous stdout result")
	}
}
