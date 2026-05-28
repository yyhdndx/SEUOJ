package queue

import "testing"

func TestNewJudgeQueue(t *testing.T) {
	if JudgeQueueKey != "judge:queue" {
		t.Fatalf("unexpected queue key %q", JudgeQueueKey)
	}
	if NewJudgeQueue(nil) == nil {
		t.Fatal("expected queue instance")
	}
}
