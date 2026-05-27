package judge

import (
	"testing"
)

func TestWorkerPureHelpers(t *testing.T) {
	if !compareOutput("1  \r\n2\t\n", "1\n2") {
		t.Fatal("expected output comparison to ignore line endings and trailing whitespace")
	}
	if compareOutput("1 2", "1  2") {
		t.Fatal("did not expect internal whitespace to be ignored")
	}
	if summarizeText("  ", 10) != "-" {
		t.Fatal("expected empty summary marker")
	}
	if got := summarizeText("abcdef", 3); got != "abc..." {
		t.Fatalf("unexpected summary %q", got)
	}
	value := 42
	if formatOptionalInt(nil) != "-" || formatOptionalInt(&value) != "42" {
		t.Fatal("unexpected optional int formatting")
	}
}

func TestNewWorkerStoresDependencies(t *testing.T) {
	worker := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil)
	if worker == nil {
		t.Fatal("expected worker instance")
	}
}
