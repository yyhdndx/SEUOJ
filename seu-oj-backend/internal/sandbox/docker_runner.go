package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var ErrInfrastructure = errors.New("sandbox infrastructure error")

type Config struct {
	Image              string
	CompileImage       string
	RunImage           string
	User               string
	CPUs               string
	MemoryMB           int
	PIDsLimit          int
	TmpfsMB            int
	FileSizeKB         int
	OutputLimitKB      int
	CompileOutputKB    int
	ReadOnlyRootFS     bool
	CompileTimeoutSec  int
	RunTimeoutBufferMS int
}

type Runner struct {
	cfg Config
}

func NewRunner(cfg Config) *Runner {
	if cfg.Image == "" {
		cfg.Image = "gcc:13"
	}
	if cfg.CompileImage == "" {
		cfg.CompileImage = cfg.Image
	}
	if cfg.RunImage == "" {
		cfg.RunImage = cfg.Image
	}
	if cfg.CPUs == "" {
		cfg.CPUs = "1.0"
	}
	if cfg.User == "" {
		cfg.User = "65534:65534"
	}
	if cfg.MemoryMB <= 0 {
		cfg.MemoryMB = 256
	}
	if cfg.PIDsLimit <= 0 {
		cfg.PIDsLimit = 64
	}
	if cfg.TmpfsMB <= 0 {
		cfg.TmpfsMB = 64
	}
	if cfg.FileSizeKB <= 0 {
		cfg.FileSizeKB = 1024
	}
	if cfg.OutputLimitKB <= 0 {
		cfg.OutputLimitKB = 256
	}
	if cfg.CompileOutputKB <= 0 {
		cfg.CompileOutputKB = 256
	}
	if cfg.CompileTimeoutSec <= 0 {
		cfg.CompileTimeoutSec = 15
	}
	if cfg.RunTimeoutBufferMS <= 0 {
		cfg.RunTimeoutBufferMS = 500
	}
	log.Printf(
		"[sandbox] compile_image=%s run_image=%s user=%s cpus=%s memory_mb=%d pids_limit=%d tmpfs_mb=%d file_size_kb=%d output_limit_kb=%d compile_output_kb=%d read_only_rootfs=%t",
		cfg.CompileImage,
		cfg.RunImage,
		cfg.User,
		cfg.CPUs,
		cfg.MemoryMB,
		cfg.PIDsLimit,
		cfg.TmpfsMB,
		cfg.FileSizeKB,
		cfg.OutputLimitKB,
		cfg.CompileOutputKB,
		cfg.ReadOnlyRootFS,
	)
	return &Runner{cfg: cfg}
}

type CompiledProgram struct {
	WorkDir  string
	RunCmd   []string
	RunImage string
}

type CompileResult struct {
	Program      *CompiledProgram
	CompileError string
}

type RunResult struct {
	Status    string
	RuntimeMS *int
	Output    string
	ErrorMsg  string
}

type languageSpec struct {
	SourceFile          string
	CompileCmd          []string
	RunCmd              []string
	DefaultCompileImage string
	DefaultRunImage     string
}

func (r *Runner) Compile(language, code string) (*CompileResult, error) {
	spec, err := r.specFor(language)
	if err != nil {
		return nil, err
	}

	workDir, err := os.MkdirTemp("", "seuoj-sandbox-*")
	if err != nil {
		return nil, err
	}

	sourcePath := filepath.Join(workDir, spec.SourceFile)
	if err := os.WriteFile(sourcePath, []byte(code), 0644); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, err
	}
	_ = os.Chmod(workDir, 0777)

	program := &CompiledProgram{
		WorkDir:  workDir,
		RunCmd:   append([]string(nil), spec.RunCmd...),
		RunImage: r.runImageFor(language, spec),
	}

	if len(spec.CompileCmd) == 0 {
		return &CompileResult{Program: program}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.cfg.CompileTimeoutSec)*time.Second)
	defer cancel()

	args := append(r.baseDockerArgs(workDir, false, true, false, r.compileImageFor(language, spec)), spec.CompileCmd...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	stderr := newLimitedBuffer(r.cfg.CompileOutputKB * 1024)
	cmd.Stderr = &stderr

	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, wrapInfrastructureError("compile timeout")
	}
	if err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderr.Exceeded() {
			stderrText += "\n[truncated]"
		}
		if isInfrastructureError(stderrText, err) {
			return nil, wrapInfrastructureError(stderrText)
		}
		return &CompileResult{
			CompileError: stderrText,
			Program:      program,
		}, nil
	}

	return &CompileResult{
		Program: program,
	}, nil
}

func (r *Runner) Run(program *CompiledProgram, input string, timeLimitMS int) RunResult {
	if program == nil || len(program.RunCmd) == 0 {
		return RunResult{
			Status:   "System Error",
			ErrorMsg: "program run command missing",
		}
	}
	if timeLimitMS <= 0 {
		timeLimitMS = 1000
	}

	timeout := time.Duration(timeLimitMS+r.cfg.RunTimeoutBufferMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := append(r.baseDockerArgs(program.WorkDir, true, false, true, program.RunImage), program.RunCmd...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(input)

	stdout := newLimitedBuffer(r.cfg.OutputLimitKB * 1024)
	stderr := newLimitedBuffer(r.cfg.OutputLimitKB * 1024)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := int(time.Since(start).Milliseconds())

	if ctx.Err() == context.DeadlineExceeded {
		return RunResult{
			Status:    "Time Limit Exceeded",
			RuntimeMS: &elapsed,
			Output:    stdout.String(),
			ErrorMsg:  "time limit exceeded",
		}
	}
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if stderr.Exceeded() {
			msg += "\n[stderr truncated]"
		}
		if isInfrastructureError(msg, err) {
			return RunResult{
				Status:    "System Error",
				RuntimeMS: &elapsed,
				Output:    stdout.String(),
				ErrorMsg:  msg,
			}
		}
		return RunResult{
			Status:    "Runtime Error",
			RuntimeMS: &elapsed,
			Output:    stdout.String(),
			ErrorMsg:  msg,
		}
	}
	if stdout.Exceeded() || stderr.Exceeded() {
		msg := "output limit exceeded"
		if stderr.Exceeded() {
			msg = "stderr output limit exceeded"
		}
		return RunResult{
			Status:    "Runtime Error",
			RuntimeMS: &elapsed,
			Output:    stdout.String(),
			ErrorMsg:  msg,
		}
	}

	return RunResult{
		Status:    "Accepted",
		RuntimeMS: &elapsed,
		Output:    stdout.String(),
	}
}

func (r *Runner) Cleanup(program *CompiledProgram) {
	if program == nil || program.WorkDir == "" {
		return
	}
	_ = os.RemoveAll(program.WorkDir)
}

func (r *Runner) Validate(ctx context.Context) error {
	versionCmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	if output, err := versionCmd.CombinedOutput(); err != nil {
		return wrapInfrastructureError(strings.TrimSpace(string(output)))
	}

	images := make(map[string]struct{})
	if strings.TrimSpace(r.cfg.Image) != "" && strings.TrimSpace(r.cfg.Image) != "gcc:13" {
		images[r.cfg.Image] = struct{}{}
	}
	if strings.TrimSpace(r.cfg.CompileImage) != "" {
		images[r.cfg.CompileImage] = struct{}{}
	}
	if strings.TrimSpace(r.cfg.RunImage) != "" {
		images[r.cfg.RunImage] = struct{}{}
	}
	for _, language := range r.SupportedLanguages() {
		spec, err := r.specFor(language)
		if err != nil {
			continue
		}
		images[r.compileImageFor(language, spec)] = struct{}{}
		images[r.runImageFor(language, spec)] = struct{}{}
	}
	log.Printf("[sandbox] validating docker images for languages=%s", strings.Join(r.SupportedLanguages(), ","))
	for image := range images {
		imageCmd := exec.CommandContext(ctx, "docker", "image", "inspect", image)
		if output, err := imageCmd.CombinedOutput(); err != nil {
			return wrapInfrastructureError(strings.TrimSpace(string(output)))
		}
	}

	return nil
}

func (r *Runner) SupportedLanguages() []string {
	return []string{"cpp", "c", "python3", "java", "go", "rust"}
}

func (r *Runner) specFor(language string) (languageSpec, error) {
	switch language {
	case "cpp":
		return languageSpec{
			SourceFile:          "main.cpp",
			CompileCmd:          []string{"g++", "-O2", "-std=c++17", "main.cpp", "-o", "main"},
			RunCmd:              []string{"./main"},
			DefaultCompileImage: "gcc:13",
			DefaultRunImage:     "gcc:13",
		}, nil
	case "c":
		return languageSpec{
			SourceFile:          "main.c",
			CompileCmd:          []string{"gcc", "-O2", "main.c", "-o", "main"},
			RunCmd:              []string{"./main"},
			DefaultCompileImage: "gcc:13",
			DefaultRunImage:     "gcc:13",
		}, nil
	case "python3":
		return languageSpec{
			SourceFile:          "main.py",
			RunCmd:              []string{"python3", "-B", "main.py"},
			DefaultCompileImage: "python:3",
			DefaultRunImage:     "python:3",
		}, nil
	case "java":
		return languageSpec{
			SourceFile:          "Main.java",
			CompileCmd:          []string{"javac", "Main.java"},
			RunCmd:              []string{"java", "Main"},
			DefaultCompileImage: "eclipse-temurin:21",
			DefaultRunImage:     "eclipse-temurin:21",
		}, nil
	case "go":
		return languageSpec{
			SourceFile:          "main.go",
			CompileCmd:          []string{"go", "build", "-o", "main", "main.go"},
			RunCmd:              []string{"./main"},
			DefaultCompileImage: "golang:1",
			DefaultRunImage:     "golang:1",
		}, nil
	case "rust":
		return languageSpec{
			SourceFile:          "main.rs",
			CompileCmd:          []string{"rustc", "-O", "main.rs", "-o", "main"},
			RunCmd:              []string{"./main"},
			DefaultCompileImage: "rust:1",
			DefaultRunImage:     "rust:1",
		}, nil
	default:
		return languageSpec{}, fmt.Errorf("unsupported language: %s", language)
	}
}

func (r *Runner) compileImageFor(language string, spec languageSpec) string {
	if strings.TrimSpace(r.cfg.CompileImage) != "" {
		return r.cfg.CompileImage
	}
	if strings.TrimSpace(r.cfg.Image) != "" && (r.cfg.Image != "gcc:13" || language == "cpp" || language == "c") {
		return r.cfg.Image
	}
	return spec.DefaultCompileImage
}

func (r *Runner) runImageFor(language string, spec languageSpec) string {
	if strings.TrimSpace(r.cfg.RunImage) != "" {
		return r.cfg.RunImage
	}
	if strings.TrimSpace(r.cfg.Image) != "" && (r.cfg.Image != "gcc:13" || language == "cpp" || language == "c") {
		return r.cfg.Image
	}
	return spec.DefaultRunImage
}

func (r *Runner) baseDockerArgs(workDir string, interactive bool, writableWorkspace bool, applyFileSizeLimit bool, image string) []string {
	workspaceMount := dockerMountPath(workDir) + ":/workspace"
	if writableWorkspace {
		workspaceMount += ":rw"
	} else {
		workspaceMount += ":ro"
	}
	args := []string{
		"run",
		"--rm",
		"--network", "none",
		"--user", r.cfg.User,
		"--cpus", r.cfg.CPUs,
		"--memory", strconv.Itoa(r.cfg.MemoryMB) + "m",
		"--memory-swap", strconv.Itoa(r.cfg.MemoryMB) + "m",
		"--pids-limit", strconv.Itoa(r.cfg.PIDsLimit),
		"--security-opt", "no-new-privileges:true",
		"--cap-drop", "ALL",
		"--ulimit", "nofile=64:64",
		"--ulimit", "nproc=64:64",
		"-e", "HOME=/tmp",
		"-v", workspaceMount,
		"-w", "/workspace",
	}
	if applyFileSizeLimit {
		args = append(args, "--ulimit", "fsize="+strconv.Itoa(r.cfg.FileSizeKB)+":"+strconv.Itoa(r.cfg.FileSizeKB))
	}
	if r.cfg.ReadOnlyRootFS {
		args = append(args, "--read-only")
	}
	args = append(args,
		"--tmpfs", "/tmp:rw,noexec,nosuid,size="+strconv.Itoa(r.cfg.TmpfsMB)+"m",
		"--tmpfs", "/run:rw,noexec,nosuid,size=16m",
	)
	if interactive {
		args = append(args, "-i")
	}
	args = append(args, image)
	return args
}

func dockerMountPath(path string) string {
	path = filepath.Clean(path)
	return strings.ReplaceAll(path, "\\", "/")
}

func isInfrastructureError(stderrText string, err error) bool {
	lower := strings.ToLower(strings.TrimSpace(stderrText))
	if lower == "" && err != nil {
		lower = strings.ToLower(err.Error())
	}
	return strings.Contains(lower, "docker:") ||
		strings.Contains(lower, "error response from daemon") ||
		strings.Contains(lower, "oci runtime create failed") ||
		strings.Contains(lower, "unable to start container process") ||
		strings.Contains(lower, "failed to create task") ||
		strings.Contains(lower, "mount denied") ||
		strings.Contains(lower, "cannot connect to the docker daemon") ||
		strings.Contains(lower, "pull access denied") ||
		strings.Contains(lower, "unable to find image")
}

func wrapInfrastructureError(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "sandbox infrastructure error"
	}
	return errors.New(ErrInfrastructure.Error() + ": " + message)
}

type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func newLimitedBuffer(limit int) limitedBuffer {
	return limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		b.exceeded = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.exceeded = true
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}

func (b *limitedBuffer) Exceeded() bool {
	return b.exceeded
}

var _ io.Writer = (*limitedBuffer)(nil)
