package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
		"# x-operon: ref=test/marker\n" +
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
	// "#:" and "x-operon:" directive lines themselves must never reach the container.
	if strings.Contains(result.Stdout, "x-operon") || strings.Contains(result.Stdout, "#:") {
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
