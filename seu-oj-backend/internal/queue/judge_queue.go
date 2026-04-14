package queue

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const JudgeQueueKey = "judge:queue"

type JudgeQueue struct {
	client *redis.Client
}

func NewJudgeQueue(client *redis.Client) *JudgeQueue {
	return &JudgeQueue{client: client}
}

func (q *JudgeQueue) EnqueueSubmission(ctx context.Context, submissionID uint64) error {
	return q.client.RPush(ctx, JudgeQueueKey, strconv.FormatUint(submissionID, 10)).Err()
}

func (q *JudgeQueue) DequeueSubmission(ctx context.Context) (uint64, error) {
	result, err := q.client.BLPop(ctx, 0, JudgeQueueKey).Result()
	if err != nil {
		return 0, err
	}
	if len(result) != 2 {
		return 0, redis.Nil
	}

	submissionID, err := strconv.ParseUint(result[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return submissionID, nil
}

func (q *JudgeQueue) Length(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, JudgeQueueKey).Result()
}
