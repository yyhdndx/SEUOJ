package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLoggingCreatesLogFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	file, err := InitLogging()
	if err != nil {
		t.Fatalf("init logging: %v", err)
	}
	defer file.Close()

	if _, err := os.Stat(filepath.Join(tempDir, "logs", "server.log")); err != nil {
		t.Fatalf("expected server log file: %v", err)
	}
}
