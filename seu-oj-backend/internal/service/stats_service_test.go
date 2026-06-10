package service

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
)

func TestStatsServiceMyAndAdmin(t *testing.T) {
	db := openServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	judgeQueue := queue.NewJudgeQueue(redisClient)
	svc := NewStatsService(db, judgeQueue, cache.New(nil))

	user := model.User{Username: "stats_user", UserID: "ST001", PasswordHash: "h", Role: "student", Status: "active"}
	problem := model.Problem{DisplayID: "ST-P1", Title: "Stats", JudgeMode: "standard", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	runtime := 10
	if err := db.Create(&[]model.Submission{
		{UserID: user.ID, ProblemID: problem.ID, Language: "cpp", Status: "Accepted", RuntimeMS: &runtime},
		{UserID: user.ID, ProblemID: problem.ID, Language: "go", Status: "Pending"},
	}).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	my, err := svc.My(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("stats my: %v", err)
	}
	if my.SubmissionsTotal != 2 || my.AcceptedSubmissions != 1 || my.PendingSubmissions != 1 || my.AcceptedProblems != 1 {
		t.Fatalf("unexpected my stats: %+v", my)
	}
	if len(my.StatusBreakdown) == 0 || len(my.LanguageBreakdown) == 0 {
		t.Fatalf("expected breakdowns: %+v", my)
	}

	if err := judgeQueue.EnqueueSubmission(context.Background(), 999); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	admin, err := svc.Admin("admin")
	if err != nil {
		t.Fatalf("stats admin: %v", err)
	}
	if admin.UsersTotal != 1 || admin.ProblemsTotal != 1 || admin.SubmissionsTotal != 2 || admin.PendingSubmissions != 1 {
		t.Fatalf("unexpected admin stats: %+v", admin)
	}
	if admin.QueueLength != 1 {
		t.Fatalf("expected queue length 1, got %d", admin.QueueLength)
	}
	if _, err := svc.Admin("student"); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected permission denied, got %v", err)
	}
}
