package testutil

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

const dockerSkipEnv = "SEUOJ_SKIP_DOCKER"

var (
	dockerChecked bool
	dockerOK      bool
	dockerMu      sync.Mutex
)

// DockerAvailable reports whether the Docker CLI can reach a running daemon.
func DockerAvailable() bool {
	dockerMu.Lock()
	defer dockerMu.Unlock()
	if dockerChecked {
		return dockerOK
	}
	dockerChecked = true
	if strings.TrimSpace(os.Getenv(dockerSkipEnv)) == "1" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	if err := cmd.Run(); err != nil {
		return false
	}
	dockerOK = true
	return true
}

// SkipUnlessDocker skips the test when Docker is unavailable or SEUOJ_SKIP_DOCKER=1.
func SkipUnlessDocker(t *testing.T) {
	t.Helper()
	if !DockerAvailable() {
		t.Skip("docker daemon not available; start Docker Desktop or set SEUOJ_SKIP_DOCKER=1 to silence")
	}
}

// EnsureDockerImage pulls the image when docker image inspect fails.
func EnsureDockerImage(t *testing.T, image string) {
	t.Helper()
	SkipUnlessDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	inspect := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	if err := inspect.Run(); err == nil {
		return
	}

	pullCtx, pullCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer pullCancel()
	t.Logf("pulling docker image %s (first run may take a while)", image)
	pull := exec.CommandContext(pullCtx, "docker", "pull", image)
	if output, err := pull.CombinedOutput(); err != nil {
		t.Fatalf("docker pull %s: %v output=%s", image, err, strings.TrimSpace(string(output)))
	}
}
