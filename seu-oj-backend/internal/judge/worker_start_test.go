package judge

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
)

func TestWorkerStartDequeuesAndStops(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db := openJudgeTestDB(t)

	problem := model.Problem{
		DisplayID: "START-1", Title: "P", Description: "d", JudgeMode: "standard",
		TimeLimitMS: 1000, MemoryLimitMB: 128, Visible: true, CreatedBy: 1,
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	submission := model.Submission{
		UserID: 1, ProblemID: problem.ID, Language: "cpp", Code: "int main(){}", Status: "Pending",
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	judgeQueue := queue.NewJudgeQueue(redisClient)
	worker := NewWorker(
		db,
		judgeQueue,
		repository.NewProblemRepository(db),
		repository.NewProblemTestcaseRepository(db),
		repository.NewSubmissionRepository(db),
		repository.NewSubmissionResultRepository(db),
		sandbox.NewRunner(sandbox.Config{}),
		cache.New(nil),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)
	if err := judgeQueue.EnqueueSubmission(context.Background(), submission.ID); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var updated model.Submission
		if err := db.First(&updated, submission.ID).Error; err != nil {
			t.Fatalf("load submission: %v", err)
		}
		if updated.Status != "Pending" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	time.Sleep(100 * time.Millisecond)
}
