package queue

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestJudgeQueueEnqueueDequeueAndLength(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	q := NewJudgeQueue(client)

	ctx := context.Background()
	if err := q.EnqueueSubmission(ctx, 101); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := q.EnqueueSubmission(ctx, 202); err != nil {
		t.Fatalf("enqueue second: %v", err)
	}

	length, err := q.Length(ctx)
	if err != nil {
		t.Fatalf("length: %v", err)
	}
	if length != 2 {
		t.Fatalf("expected queue length 2, got %d", length)
	}

	first, err := q.DequeueSubmission(ctx)
	if err != nil {
		t.Fatalf("dequeue first: %v", err)
	}
	if first != 101 {
		t.Fatalf("expected first submission 101, got %d", first)
	}

	second, err := q.DequeueSubmission(ctx)
	if err != nil {
		t.Fatalf("dequeue second: %v", err)
	}
	if second != 202 {
		t.Fatalf("expected second submission 202, got %d", second)
	}

	length, err = q.Length(ctx)
	if err != nil {
		t.Fatalf("length after dequeue: %v", err)
	}
	if length != 0 {
		t.Fatalf("expected empty queue, got %d", length)
	}
}

func TestJudgeQueueDequeueEmptyBlocksUntilCancelled(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	q := NewJudgeQueue(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := q.DequeueSubmission(ctx)
		done <- err
	}()

	if err := q.EnqueueSubmission(context.Background(), 7); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("dequeue: %v", err)
	}
}

func TestJudgeQueueDequeueInvalidPayload(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	if err := client.RPush(context.Background(), JudgeQueueKey, "not-a-number").Err(); err != nil {
		t.Fatalf("seed invalid payload: %v", err)
	}

	q := NewJudgeQueue(client)
	_, err := q.DequeueSubmission(context.Background())
	if err == nil {
		t.Fatal("expected parse error for invalid payload")
	}
}

func TestJudgeQueueDequeueCanceledContext(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	q := NewJudgeQueue(client)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := q.DequeueSubmission(ctx)
		done <- err
	}()

	cancel()
	if err := <-done; err == nil {
		t.Fatal("expected canceled context error")
	}
}
