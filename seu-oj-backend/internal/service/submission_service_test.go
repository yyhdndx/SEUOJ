package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
)

func newTestSubmissionService(t *testing.T, redisClient *redis.Client) *SubmissionService {
	t.Helper()

	db := openServiceTestDB(t)
	problemRepo := repository.NewProblemRepository(db)
	testcaseRepo := repository.NewProblemTestcaseRepository(db)
	contestService := NewContestService(db, problemRepo, testcaseRepo, cache.New(nil))
	return NewSubmissionService(
		db,
		repository.NewSubmissionRepository(db),
		repository.NewSubmissionResultRepository(db),
		problemRepo,
		testcaseRepo,
		queue.NewJudgeQueue(redisClient),
		sandbox.NewRunner(sandbox.Config{}),
		contestService,
		cache.New(nil),
	)
}

func seedVisibleProblem(t *testing.T, svc *SubmissionService, displayID string) uint64 {
	t.Helper()

	problem := model.Problem{
		DisplayID:     displayID,
		Title:         "Test Problem",
		Description:   "desc",
		JudgeMode:     "standard",
		TimeLimitMS:   1000,
		MemoryLimitMB: 128,
		Visible:       true,
		CreatedBy:     1,
	}
	if err := svc.problemRepo.Create(svc.db, &problem); err != nil {
		t.Fatalf("create problem: %v", err)
	}
	return problem.ID
}

func TestSubmissionServiceCreateSubmissionEnqueues(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-001")

	id, status, err := svc.CreateSubmission(1, "student", dto.CreateSubmissionRequest{
		ProblemID: problemID,
		Language:  "cpp",
		Code:      "#include <iostream>\nint main(){}",
	})
	if err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if status != "Pending" || id == 0 {
		t.Fatalf("unexpected create result id=%d status=%q", id, status)
	}

	length, err := svc.judgeQueue.Length(context.Background())
	if err != nil {
		t.Fatalf("queue length: %v", err)
	}
	if length != 1 {
		t.Fatalf("expected one queued submission, got %d", length)
	}
}

func TestSubmissionServiceCreateSubmissionRejectsHiddenProblem(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))

	problem := model.Problem{
		DisplayID:     "SUB-HIDDEN",
		Title:         "Hidden",
		Description:   "desc",
		JudgeMode:     "standard",
		TimeLimitMS:   1000,
		MemoryLimitMB: 128,
		Visible:       false,
		CreatedBy:     1,
	}
	if err := svc.problemRepo.Create(svc.db, &problem); err != nil {
		t.Fatalf("create hidden problem: %v", err)
	}

	_, _, err := svc.CreateSubmission(1, "student", dto.CreateSubmissionRequest{
		ProblemID: problem.ID,
		Language:  "cpp",
		Code:      "int main(){}",
	})
	if !errors.Is(err, ErrProblemUnavailable) {
		t.Fatalf("expected hidden problem error, got %v", err)
	}
}

func TestSubmissionServiceEnqueueFailureMarksSystemError(t *testing.T) {
	brokenRedis := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = brokenRedis.Close() })

	svc := newTestSubmissionService(t, brokenRedis)
	problemID := seedVisibleProblem(t, svc, "SUB-ERR")

	_, _, err := svc.CreateSubmission(2, "student", dto.CreateSubmissionRequest{
		ProblemID: problemID,
		Language:  "cpp",
		Code:      "int main(){}",
	})
	if !errors.Is(err, ErrQueueEnqueueFailed) {
		t.Fatalf("expected enqueue failed, got %v", err)
	}

	var count int64
	if err := svc.db.Model(&model.Submission{}).Where("user_id = ? AND status = ?", 2, "System Error").Count(&count).Error; err != nil {
		t.Fatalf("count system error submissions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one system error submission, got %d", count)
	}
}

func TestSubmissionServiceListMySubmissionsUsesRealTotal(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-PAGE")

	for i := 0; i < 25; i++ {
		submission := model.Submission{
			UserID:    9,
			ProblemID: problemID,
			Language:  "cpp",
			Code:      "int main(){}",
			Status:    "Accepted",
		}
		if err := svc.submissionRepo.Create(&submission); err != nil {
			t.Fatalf("seed submission %d: %v", i, err)
		}
	}

	resp, err := svc.ListMySubmissions(context.Background(), 9, 1, 20, nil, nil, nil)
	if err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if resp.Total != 25 {
		t.Fatalf("expected total=25, got %d", resp.Total)
	}
	if len(resp.List) != 20 {
		t.Fatalf("expected page size 20, got %d", len(resp.List))
	}
}

func TestSubmissionServiceRejudgeRequiresAdmin(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-REJ")

	submission := model.Submission{
		UserID:    1,
		ProblemID: problemID,
		Language:  "cpp",
		Code:      "int main(){}",
		Status:    "Wrong Answer",
	}
	if err := svc.submissionRepo.Create(&submission); err != nil {
		t.Fatalf("create submission: %v", err)
	}

	_, _, err := svc.RejudgeSubmission("student", submission.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected permission denied, got %v", err)
	}

	id, status, err := svc.RejudgeSubmission("admin", submission.ID)
	if err != nil {
		t.Fatalf("rejudge: %v", err)
	}
	if id != submission.ID || status != "Pending" {
		t.Fatalf("unexpected rejudge result id=%d status=%q", id, status)
	}
}

func TestSubmissionServiceGetSubmissionDetail(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-DET")

	submission := model.Submission{
		UserID: 3, ProblemID: problemID, Language: "cpp", Code: "int main(){}",
		Status: "Accepted", PassedCount: 1, TotalCount: 1,
	}
	if err := svc.submissionRepo.Create(&submission); err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if err := svc.submissionResultRepo.ReplaceBySubmissionID(svc.db, submission.ID, []model.SubmissionResult{
		{SubmissionID: submission.ID, TestcaseID: 1, Status: "Accepted"},
	}); err != nil {
		t.Fatalf("create results: %v", err)
	}

	detail, err := svc.GetSubmissionDetail(context.Background(), 3, "student", submission.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.ID != submission.ID || detail.Status != "Accepted" || len(detail.Results) != 1 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	if _, err := svc.GetSubmissionDetail(context.Background(), 99, "student", submission.ID); !errors.Is(err, ErrSubmissionForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}

	adminDetail, err := svc.GetSubmissionDetail(context.Background(), 1, "admin", submission.ID)
	if err != nil || adminDetail.ID != submission.ID {
		t.Fatalf("admin detail: err=%v detail=%+v", err, adminDetail)
	}
}

func TestSubmissionServiceListPublicAndFiltered(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-PUB")

	user := model.User{Username: "pubuser", UserID: "PUB01", PasswordHash: "h", Role: "student", Status: "active"}
	if err := svc.db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	for i := 0; i < 3; i++ {
		status := "Wrong Answer"
		if i == 2 {
			status = "Accepted"
		}
		if err := svc.submissionRepo.Create(&model.Submission{
			UserID: user.ID, ProblemID: problemID, Language: "cpp", Code: "x", Status: status,
		}); err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}

	public, err := svc.ListPublicSubmissions(1, 10, nil, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("list public: %v", err)
	}
	if public.Total != 3 || len(public.List) != 3 {
		t.Fatalf("unexpected public list: %+v", public)
	}

	accepted := "Accepted"
	filtered, err := svc.ListPublicSubmissions(1, 10, &user.ID, &problemID, nil, &accepted, "cpp")
	if err != nil {
		t.Fatalf("list public filtered: %v", err)
	}
	if filtered.Total != 1 || filtered.List[0].Status != "Accepted" {
		t.Fatalf("unexpected filtered public list: %+v", filtered)
	}

	status := "Accepted"
	myList, err := svc.ListMySubmissions(context.Background(), user.ID, 1, 10, &problemID, nil, &status)
	if err != nil {
		t.Fatalf("list my filtered: %v", err)
	}
	if myList.Total != 1 {
		t.Fatalf("expected filtered total 1, got %+v", myList)
	}

	adminList, err := svc.ListSubmissions("admin", 1, 10, &user.ID, nil, nil, nil)
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if adminList.Total != 3 {
		t.Fatalf("unexpected admin list: %+v", adminList)
	}
	if _, err := svc.ListSubmissions("student", 1, 10, nil, nil, nil, nil); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected admin permission error, got %v", err)
	}
}

func TestSubmissionServiceListRecentMySubmissions(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := newTestSubmissionService(t, redisClient)
	svc.cache = cache.New(redisClient)
	problemID := seedVisibleProblem(t, svc, "SUB-REC")

	for i := 0; i < 5; i++ {
		if err := svc.submissionRepo.Create(&model.Submission{
			UserID: 4, ProblemID: problemID, Language: "cpp", Code: "x", Status: "Accepted",
		}); err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}

	resp, err := svc.ListMySubmissions(context.Background(), 4, 1, 3, nil, nil, nil)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(resp.List) != 3 || resp.Total != 5 {
		t.Fatalf("unexpected recent list: total=%d len=%d", resp.Total, len(resp.List))
	}
}

func TestRunSampleTestsNoSampleTestcases(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "RUN-NOSAMPLE")

	resp, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problemID, Language: "cpp", Code: "int main(){}",
	})
	if err != nil {
		t.Fatalf("run sample tests: %v", err)
	}
	if resp.Status != "System Error" || resp.ErrorMsg != "no sample testcases" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSubmissionServiceCreateSubmissionInContest(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "SUB-CT")

	now := time.Now()
	contestID, err := svc.contestService.Create(1, "admin", dto.CreateContestRequest{
		Title: "Mini", RuleType: "acm",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour),
		IsPublic: true, AllowPractice: true,
		Problems: []dto.CreateContestProblemDTO{{ProblemID: problemID}},
	})
	if err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := svc.contestService.Register(2, "student", contestID); err != nil {
		t.Fatalf("register: %v", err)
	}

	id, status, err := svc.CreateSubmission(2, "student", dto.CreateSubmissionRequest{
		ProblemID: problemID, ContestID: &contestID, Language: "cpp", Code: "int main(){}",
	})
	if err != nil {
		t.Fatalf("contest submission: %v", err)
	}
	if id == 0 || status != "Pending" {
		t.Fatalf("unexpected submission id=%d status=%q", id, status)
	}

	contestIDCopy := contestID
	resp, err := svc.RunSampleTests(2, "student", dto.RunSubmissionRequest{
		ProblemID: problemID, ContestID: &contestIDCopy, Language: "cpp", Code: "int main(){}",
	})
	if err != nil {
		t.Fatalf("run in contest: %v", err)
	}
	if resp.Status != "System Error" {
		t.Fatalf("expected no sample testcases system error, got %+v", resp)
	}
}

func TestRunSampleTestsHiddenProblem(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problem := model.Problem{
		DisplayID: "RUN-HID", Title: "Hidden", Description: "d", JudgeMode: "standard",
		TimeLimitMS: 1000, MemoryLimitMB: 128, Visible: false, CreatedBy: 1,
	}
	if err := svc.problemRepo.Create(svc.db, &problem); err != nil {
		t.Fatalf("create problem: %v", err)
	}
	_, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problem.ID, Language: "cpp", Code: "int main(){}",
	})
	if !errors.Is(err, ErrProblemUnavailable) {
		t.Fatalf("expected hidden problem error, got %v", err)
	}
}
