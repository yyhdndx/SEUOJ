package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"seu-oj-backend/internal/testutil"
)

func dockerPythonRunner(t *testing.T) *Runner {
	t.Helper()
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")
	return NewRunner(Config{
		Image:        "python:3",
		CompileImage: "python:3",
		RunImage:     "python:3",
	})
}

func dockerCppRunner(t *testing.T) *Runner {
	t.Helper()
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "gcc:13")
	return NewRunner(Config{Image: "gcc:13"})
}

func TestDockerPythonRunAccepted(t *testing.T) {
	runner := dockerPythonRunner(t)
	code := "a, b = map(int, input().split())\nprint(a + b)\n"

	compiled, err := runner.Compile("python3", code)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	defer runner.Cleanup(compiled.Program)
	if compiled.CompileError != "" {
		t.Fatalf("unexpected compile error: %s", compiled.CompileError)
	}

	result := runner.Run(compiled.Program, "2 3\n", 2000)
	if result.Status != "Accepted" {
		t.Fatalf("expected Accepted, got %s err=%q output=%q", result.Status, result.ErrorMsg, result.Output)
	}
	if strings.TrimSpace(result.Output) != "5" {
		t.Fatalf("unexpected output %q", result.Output)
	}
	if result.RuntimeMS == nil || *result.RuntimeMS <= 0 {
		t.Fatalf("expected runtime_ms, got %+v", result.RuntimeMS)
	}
}

func TestDockerPythonWrongAnswer(t *testing.T) {
	runner := dockerPythonRunner(t)
	code := "print(int(input()) + 1)\n"

	compiled, err := runner.Compile("python3", code)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	defer runner.Cleanup(compiled.Program)

	result := runner.Run(compiled.Program, "3\n", 2000)
	if result.Status != "Accepted" {
		t.Fatalf("sandbox run status=%s err=%q", result.Status, result.ErrorMsg)
	}
	if strings.TrimSpace(result.Output) == "3" {
		t.Fatal("expected wrong output for WA scenario")
	}
}

func TestDockerPythonRuntimeError(t *testing.T) {
	runner := dockerPythonRunner(t)
	code := "raise RuntimeError('boom')\n"

	compiled, err := runner.Compile("python3", code)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	defer runner.Cleanup(compiled.Program)

	result := runner.Run(compiled.Program, "", 2000)
	if result.Status != "Runtime Error" {
		t.Fatalf("expected Runtime Error, got %s err=%q", result.Status, result.ErrorMsg)
	}
}

func TestDockerPythonTimeLimitExceeded(t *testing.T) {
	runner := dockerPythonRunner(t)
	code := "while True:\n    pass\n"

	compiled, err := runner.Compile("python3", code)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	defer runner.Cleanup(compiled.Program)

	result := runner.Run(compiled.Program, "", 200)
	if result.Status != "Time Limit Exceeded" {
		t.Fatalf("expected TLE, got %s err=%q", result.Status, result.ErrorMsg)
	}
}

func TestDockerCppCompileAndRunAccepted(t *testing.T) {
	runner := dockerCppRunner(t)
	code := `#include <bits/stdc++.h>
using namespace std;
int main() {
    long long a, b;
    if (!(cin >> a >> b)) return 0;
    cout << a + b << "\n";
    return 0;
}`

	compiled, err := runner.Compile("cpp", code)
	if err != nil {
		t.Fatalf("compile command: %v", err)
	}
	defer runner.Cleanup(compiled.Program)
	if compiled.CompileError != "" {
		t.Fatalf("unexpected compile error: %s", compiled.CompileError)
	}

	result := runner.Run(compiled.Program, "7 8\n", 2000)
	if result.Status != "Accepted" || strings.TrimSpace(result.Output) != "15" {
		t.Fatalf("unexpected run result status=%s output=%q err=%q", result.Status, result.Output, result.ErrorMsg)
	}
}

func TestDockerCppCompileError(t *testing.T) {
	runner := dockerCppRunner(t)

	compiled, err := runner.Compile("cpp", "int main() { return nope; }")
	if err != nil {
		t.Fatalf("compile command: %v", err)
	}
	defer runner.Cleanup(compiled.Program)
	if strings.TrimSpace(compiled.CompileError) == "" {
		t.Fatal("expected compile error output")
	}
}

func TestDockerValidatePythonImage(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")

	runner := NewRunner(Config{
		Image:        "python:3",
		CompileImage: "python:3",
		RunImage:     "python:3",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := runner.Validate(ctx); err != nil {
		t.Fatalf("validate python sandbox: %v", err)
	}
}
