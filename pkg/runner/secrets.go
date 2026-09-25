package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/ofthemachine/fraglet/pkg/dockercli"
)

// secretContainer is a container that was created (not run) so its secret
// files could be copied in before it starts: the same create → copy → start
// sequence Docker Compose uses for environment-sourced secrets. The values
// travel as an in-memory tar on `docker cp -`'s stdin, so they never appear
// in argv, in `docker inspect`, in the container's environment, or in a host
// file. The container is created with --rm and removes itself when it exits;
// fragletc also force-removes it on an interrupt or once the run is over, so
// a secret outlives the run only if fragletc itself is SIGKILLed.
type secretContainer struct {
	id      string
	signals chan os.Signal
	stop    chan struct{}
	once    sync.Once
}

// createWithSecrets runs createArgs (a full "docker create ..." argv), copies
// each secret into the new container, and returns it ready to start. From
// here until release, SIGINT and SIGTERM force-remove the container instead
// of killing fragletc, so an interrupt at any point removes the secret too.
func createWithSecrets(ctx context.Context, createArgs []string, secrets []Secret) (*secretContainer, error) {
	archive, err := secretsTar(secrets)
	if err != nil {
		return nil, err
	}

	sc := &secretContainer{signals: make(chan os.Signal, 1), stop: make(chan struct{})}
	signal.Notify(sc.signals, os.Interrupt, syscall.SIGTERM)

	var stdout, stderr bytes.Buffer
	create := exec.CommandContext(ctx, createArgs[0], createArgs[1:]...)
	create.Stdout, create.Stderr = &stdout, &stderr
	if err := create.Run(); err != nil {
		signal.Stop(sc.signals)
		return nil, fmt.Errorf("docker create: %v\n%s", err, stderr.String())
	}
	sc.id = strings.TrimSpace(stdout.String())

	cp := exec.CommandContext(ctx, dockercli.Binary(), "cp", "-", sc.id+":/")
	cp.Stdin = bytes.NewReader(archive)
	if out, err := cp.CombinedOutput(); err != nil {
		sc.release()
		return nil, fmt.Errorf("docker cp (secrets): %v\n%s", err, out)
	}

	select {
	case s := <-sc.signals:
		sc.release()
		return nil, fmt.Errorf("interrupted (%v) before the container started; it was removed", s)
	default:
	}
	go func() {
		select {
		case <-sc.signals:
			sc.remove()
		case <-sc.stop:
		}
	}()
	return sc, nil
}

// startArgs is the argv that starts the container attached, forwarding stdin
// when the create had -i. `docker start -a` exits with the container's code.
func (sc *secretContainer) startArgs(attachStdin bool) []string {
	args := []string{dockercli.Binary(), "start", "-a"}
	if attachStdin {
		args = append(args, "-i")
	}
	return append(args, sc.id)
}

// remove force-removes the container. It is harmless after --rm already has.
func (sc *secretContainer) remove() {
	_ = exec.Command(dockercli.Binary(), "rm", "-f", sc.id).Run()
}

// release ends the interrupt watch and removes the container. Idempotent.
func (sc *secretContainer) release() {
	sc.once.Do(func() {
		signal.Stop(sc.signals)
		close(sc.stop)
		sc.remove()
	})
}

// secretsTar builds the archive `docker cp - <id>:/` extracts: one read-only
// (0444, root-owned) file per secret at its container path. Parent
// directories are created by docker cp, so none are listed, which keeps an
// image's existing /run untouched.
func secretsTar(secrets []Secret) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, s := range secrets {
		name := strings.TrimPrefix(s.Path, "/")
		if name == "" || name == s.Path {
			return nil, fmt.Errorf("secret path %q must be absolute", s.Path)
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o444, Size: int64(len(s.Value)), Typeflag: tar.TypeReg}); err != nil {
			return nil, err
		}
		if _, err := tw.Write(s.Value); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
