package sandbox

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRunnerAppliesDefaults(t *testing.T) {
	runner := NewRunner(Config{})

	if runner.cfg.Image != "gcc:13" || runner.cfg.CompileImage != "" || runner.cfg.RunImage != "" {
		t.Fatalf("unexpected image defaults: %+v", runner.cfg)
	}
	if runner.cfg.User != "65534:65534" || runner.cfg.CPUs != "1.0" {
		t.Fatalf("unexpected execution defaults: %+v", runner.cfg)
	}
	if runner.cfg.MemoryMB != 256 || runner.cfg.PIDsLimit != 64 || runner.cfg.TmpfsMB != 64 {
		t.Fatalf("unexpected resource defaults: %+v", runner.cfg)
	}
	if runner.cfg.FileSizeKB != 1024 || runner.cfg.OutputLimitKB != 256 || runner.cfg.CompileOutputKB != 256 {
		t.Fatalf("unexpected IO defaults: %+v", runner.cfg)
	}
}

func TestLanguageSpecsAndImageSelection(t *testing.T) {
	runner := NewRunner(Config{})
	for _, language := range runner.SupportedLanguages() {
		spec, err := runner.specFor(language)
		if err != nil {
			t.Fatalf("spec for %s: %v", language, err)
		}
		if spec.SourceFile == "" || len(spec.RunCmd) == 0 {
			t.Fatalf("incomplete spec for %s: %+v", language, spec)
		}
	}
	if _, err := runner.specFor("pascal"); err == nil {
		t.Fatal("expected unsupported language error")
	}

	goSpec, _ := runner.specFor("go")
	if got := runner.compileImageFor("go", goSpec); got != "golang:1" {
		t.Fatalf("expected go compile image, got %q", got)
	}
	cppSpec, _ := runner.specFor("cpp")
	if got := runner.runImageFor("cpp", cppSpec); got != "gcc:13" {
		t.Fatalf("expected cpp run image, got %q", got)
	}

	custom := NewRunner(Config{CompileImage: "build:latest", RunImage: "run:latest"})
	if got := custom.compileImageFor("go", goSpec); got != "build:latest" {
		t.Fatalf("expected custom compile image, got %q", got)
	}
	if got := custom.runImageFor("go", goSpec); got != "run:latest" {
		t.Fatalf("expected custom run image, got %q", got)
	}
}

func TestBaseDockerArgs(t *testing.T) {
	runner := NewRunner(Config{
		Image:          "gcc:13",
		User:           "1000:1000",
		CPUs:           "0.5",
		MemoryMB:       128,
		PIDsLimit:      32,
		TmpfsMB:        16,
		FileSizeKB:     64,
		ReadOnlyRootFS: true,
	})

	args := runner.baseDockerArgs(`C:\tmp\work`, true, false, true, "gcc:13")
	joined := strings.Join(args, " ")
	for _, expected := range []string{
		"--network none",
		"--user 1000:1000",
		"--cpus 0.5",
		"--memory 128m",
		"--pids-limit 32",
		"--ulimit fsize=64:64",
		"--read-only",
		"-i",
		"gcc:13",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected docker args to contain %q in %q", expected, joined)
		}
	}
	if !strings.Contains(joined, "C:/tmp/work:/workspace:ro") {
		t.Fatalf("expected normalized read-only mount, got %q", joined)
	}
}

func TestLimitedBufferTracksTruncation(t *testing.T) {
	buffer := newLimitedBuffer(5)
	n, err := buffer.Write([]byte("abcdef"))
	if err != nil || n != 6 {
		t.Fatalf("unexpected write result n=%d err=%v", n, err)
	}
	if buffer.String() != "abcde" || !buffer.Exceeded() {
		t.Fatalf("expected truncated buffer, got %q exceeded=%t", buffer.String(), buffer.Exceeded())
	}

	disabled := newLimitedBuffer(0)
	n, err = disabled.Write([]byte("ignored"))
	if err != nil || n != len("ignored") || disabled.String() != "" || disabled.Exceeded() {
		t.Fatalf("unexpected disabled buffer state n=%d err=%v value=%q exceeded=%t", n, err, disabled.String(), disabled.Exceeded())
	}
}

func TestInfrastructureErrorDetection(t *testing.T) {
	if !isInfrastructureError("Cannot connect to the Docker daemon", nil) {
		t.Fatal("expected docker daemon error to be infrastructure")
	}
	if isInfrastructureError("program returned non-zero", errors.New("exit status 1")) {
		t.Fatal("did not expect regular runtime failure to be infrastructure")
	}
	err := wrapInfrastructureError("")
	if !strings.Contains(err.Error(), ErrInfrastructure.Error()) {
		t.Fatalf("expected wrapped infrastructure error, got %v", err)
	}
}
