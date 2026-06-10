package judge

import (
	"context"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
	"seu-oj-backend/internal/testutil"
)

func openJudgeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(
		&model.User{},
		&model.Problem{},
		&model.ProblemTestcase{},
		&model.Submission{},
		&model.SubmissionResult{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newWorkerForTest(t *testing.T, db *gorm.DB, runner *sandbox.Runner) *Worker {
	t.Helper()

	return NewWorker(
		db,
		queue.NewJudgeQueue(nil),
		repository.NewProblemRepository(db),
		repository.NewProblemTestcaseRepository(db),
		repository.NewSubmissionRepository(db),
		repository.NewSubmissionResultRepository(db),
		runner,
		cache.New(nil),
	)
}

func seedJudgeProblem(t *testing.T, db *gorm.DB, input, output string) (uint64, uint64) {
	t.Helper()

	problem := model.Problem{
		DisplayID:     "JUDGE-1",
		Title:         "Add",
		Description:   "add",
		JudgeMode:     "standard",
		TimeLimitMS:   2000,
		MemoryLimitMB: 128,
		Visible:       true,
		CreatedBy:     1,
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	testcase := model.ProblemTestcase{
		ProblemID:  problem.ID,
		CaseType:   "sample",
		InputData:  input,
		OutputData: output,
		SortOrder:  1,
		IsActive:   true,
		Score:      100,
	}
	if err := db.Create(&testcase).Error; err != nil {
		t.Fatalf("create testcase: %v", err)
	}
	return problem.ID, testcase.ID
}

func seedPendingSubmission(t *testing.T, db *gorm.DB, problemID uint64, language, code string) uint64 {
	t.Helper()

	submission := model.Submission{
		UserID:    1,
		ProblemID: problemID,
		Language:  language,
		Code:      code,
		Status:    "Pending",
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}
	return submission.ID
}

func TestWorkerHandleSubmissionSkipsNonPending(t *testing.T) {
	db := openJudgeTestDB(t)
	worker := newWorkerForTest(t, db, sandbox.NewRunner(sandbox.Config{}))

	problemID, _ := seedJudgeProblem(t, db, "1\n", "1\n")
	submissionID := seedPendingSubmission(t, db, problemID, "python3", "print(1)")
	if err := db.Model(&model.Submission{}).Where("id = ?", submissionID).Update("status", "Accepted").Error; err != nil {
		t.Fatalf("update status: %v", err)
	}

	if err := worker.handleSubmission(context.Background(), submissionID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}
	var submission model.Submission
	if err := db.First(&submission, submissionID).Error; err != nil {
		t.Fatalf("load submission: %v", err)
	}
	if submission.Status != "Accepted" {
		t.Fatalf("expected status unchanged, got %q", submission.Status)
	}
}

func TestWorkerHandleSubmissionProblemNotFound(t *testing.T) {
	db := openJudgeTestDB(t)
	worker := newWorkerForTest(t, db, sandbox.NewRunner(sandbox.Config{}))

	submission := model.Submission{
		UserID: 1, ProblemID: 999, Language: "python3", Code: "print(1)", Status: "Pending",
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	if err := worker.handleSubmission(context.Background(), submission.ID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}
	if err := db.First(&submission, submission.ID).Error; err != nil {
		t.Fatalf("reload submission: %v", err)
	}
	if submission.Status != "System Error" {
		t.Fatalf("expected System Error, got %q msg=%q", submission.Status, submission.ErrorMsg)
	}
}

func TestWorkerHandleSubmissionNoActiveTestcases(t *testing.T) {
	db := openJudgeTestDB(t)
	worker := newWorkerForTest(t, db, sandbox.NewRunner(sandbox.Config{}))

	problem := model.Problem{
		DisplayID: "JUDGE-NOCASE", Title: "No Case", Description: "d", JudgeMode: "standard",
		TimeLimitMS: 1000, MemoryLimitMB: 128, Visible: true, CreatedBy: 1,
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	if err := db.Create(&model.ProblemTestcase{
		ProblemID: problem.ID, CaseType: "hidden", InputData: "1", OutputData: "1",
		SortOrder: 1, IsActive: false, Score: 100,
	}).Error; err != nil {
		t.Fatalf("create inactive testcase: %v", err)
	}

	submissionID := seedPendingSubmission(t, db, problem.ID, "python3", "print(1)")
	if err := worker.handleSubmission(context.Background(), submissionID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}
	var submission model.Submission
	if err := db.First(&submission, submissionID).Error; err != nil {
		t.Fatalf("reload submission: %v", err)
	}
	if submission.Status != "System Error" || submission.ErrorMsg != "no active testcases" {
		t.Fatalf("unexpected submission after no testcase: status=%q msg=%q", submission.Status, submission.ErrorMsg)
	}
}

func TestWorkerHandleSubmissionAcceptedWithDocker(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")

	db := openJudgeTestDB(t)
	runner := sandbox.NewRunner(sandbox.Config{Image: "python:3", CompileImage: "python:3", RunImage: "python:3"})
	worker := newWorkerForTest(t, db, runner)

	problemID, testcaseID := seedJudgeProblem(t, db, "4 5\n", "9\n")
	code := "a, b = map(int, input().split())\nprint(a + b)\n"
	submissionID := seedPendingSubmission(t, db, problemID, "python3", code)

	if err := worker.handleSubmission(context.Background(), submissionID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}

	var submission model.Submission
	if err := db.First(&submission, submissionID).Error; err != nil {
		t.Fatalf("reload submission: %v", err)
	}
	if submission.Status != "Accepted" || submission.PassedCount != 1 || submission.TotalCount != 1 {
		t.Fatalf("unexpected submission result: %+v", submission)
	}

	var results []model.SubmissionResult
	if err := db.Where("submission_id = ?", submissionID).Find(&results).Error; err != nil {
		t.Fatalf("load results: %v", err)
	}
	if len(results) != 1 || results[0].TestcaseID != testcaseID || results[0].Status != "Accepted" {
		t.Fatalf("unexpected testcase results: %+v", results)
	}
}

func TestWorkerHandleSubmissionWrongAnswerWithDocker(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")

	db := openJudgeTestDB(t)
	runner := sandbox.NewRunner(sandbox.Config{Image: "python:3", CompileImage: "python:3", RunImage: "python:3"})
	worker := newWorkerForTest(t, db, runner)

	problemID, _ := seedJudgeProblem(t, db, "2\n", "2\n")
	submissionID := seedPendingSubmission(t, db, problemID, "python3", "print(3)\n")

	if err := worker.handleSubmission(context.Background(), submissionID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}

	var submission model.Submission
	if err := db.First(&submission, submissionID).Error; err != nil {
		t.Fatalf("reload submission: %v", err)
	}
	if submission.Status != "Wrong Answer" || submission.PassedCount != 0 {
		t.Fatalf("unexpected WA submission: %+v", submission)
	}
}

func TestWorkerHandleSubmissionCompileErrorWithDocker(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "gcc:13")

	db := openJudgeTestDB(t)
	worker := newWorkerForTest(t, db, sandbox.NewRunner(sandbox.Config{Image: "gcc:13"}))

	problemID, _ := seedJudgeProblem(t, db, "1\n", "1\n")
	submissionID := seedPendingSubmission(t, db, problemID, "cpp", "int main() { return nope; }")

	if err := worker.handleSubmission(context.Background(), submissionID); err != nil {
		t.Fatalf("handle submission: %v", err)
	}

	var submission model.Submission
	if err := db.First(&submission, submissionID).Error; err != nil {
		t.Fatalf("reload submission: %v", err)
	}
	if submission.Status != "Compile Error" || strings.TrimSpace(submission.CompileInfo) == "" {
		t.Fatalf("unexpected compile error submission: %+v", submission)
	}
}
