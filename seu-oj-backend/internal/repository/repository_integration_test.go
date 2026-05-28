package repository

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/model"
)

func openRepositoryTestDB(t *testing.T) *gorm.DB {
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
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&model.Problem{},
		&model.ProblemTestcase{},
		&model.Submission{},
		&model.SubmissionResult{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestProblemRepositoryCRUDAndListFilters(t *testing.T) {
	db := openRepositoryTestDB(t)
	repo := NewProblemRepository(db)

	problems := []model.Problem{
		{DisplayID: "A100", Title: "A plus B", JudgeMode: "standard", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128},
		{DisplayID: "B200", Title: "Hidden Sum", JudgeMode: "standard", Visible: false, TimeLimitMS: 1000, MemoryLimitMB: 128},
		{DisplayID: "C300", Title: "Visible Sum", JudgeMode: "standard", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128},
	}
	for i := range problems {
		if err := repo.Create(db, &problems[i]); err != nil {
			t.Fatalf("create problem %d: %v", i, err)
		}
	}

	visible, total, err := repo.ListVisible(1, 10, "Sum")
	if err != nil {
		t.Fatalf("list visible: %v", err)
	}
	if total != 1 || len(visible) != 1 || visible[0].DisplayID != "C300" {
		t.Fatalf("unexpected visible list total=%d list=%+v", total, visible)
	}

	all, total, err := repo.List(1, 10, "", true)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if total != 3 || len(all) != 3 || all[0].DisplayID != "C300" {
		t.Fatalf("unexpected all list total=%d list=%+v", total, all)
	}

	problem, err := repo.GetByID(problems[0].ID)
	if err != nil {
		t.Fatalf("get problem: %v", err)
	}
	problem.Title = "Updated"
	if err := repo.Update(db, problem); err != nil {
		t.Fatalf("update problem: %v", err)
	}
	updated, err := repo.GetByID(problem.ID)
	if err != nil {
		t.Fatalf("get updated problem: %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	if err := repo.Delete(db, problem.ID); err != nil {
		t.Fatalf("delete problem: %v", err)
	}
	if _, err := repo.GetByID(problem.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}

func TestProblemTestcaseRepositoryOrdersAndReplaces(t *testing.T) {
	db := openRepositoryTestDB(t)
	problem := model.Problem{DisplayID: "P1", Title: "Problem", JudgeMode: "standard", Visible: true}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}

	repo := NewProblemTestcaseRepository(db)
	initial := []model.ProblemTestcase{
		{ProblemID: problem.ID, CaseType: "hidden", InputData: "2", OutputData: "2", SortOrder: 2, IsActive: true},
		{ProblemID: problem.ID, CaseType: "sample", InputData: "1", OutputData: "1", SortOrder: 1, IsActive: true},
	}
	if err := repo.BatchCreate(db, initial); err != nil {
		t.Fatalf("batch create: %v", err)
	}

	list, err := repo.ListByProblemID(problem.ID)
	if err != nil {
		t.Fatalf("list testcases: %v", err)
	}
	if len(list) != 2 || list[0].SortOrder != 1 || list[1].SortOrder != 2 {
		t.Fatalf("unexpected testcase order: %+v", list)
	}

	replacement := []model.ProblemTestcase{
		{ProblemID: problem.ID, CaseType: "sample", InputData: "3", OutputData: "3", SortOrder: 3, IsActive: true},
	}
	if err := repo.ReplaceByProblemID(db, problem.ID, replacement); err != nil {
		t.Fatalf("replace testcases: %v", err)
	}
	list, err = repo.ListByProblemID(problem.ID)
	if err != nil {
		t.Fatalf("list replacement: %v", err)
	}
	if len(list) != 1 || list[0].InputData != "3" {
		t.Fatalf("unexpected replacement list: %+v", list)
	}

	if err := repo.ReplaceByProblemID(db, problem.ID, nil); err != nil {
		t.Fatalf("replace empty: %v", err)
	}
	list, err = repo.ListByProblemID(problem.ID)
	if err != nil {
		t.Fatalf("list empty replacement: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected all testcases removed, got %+v", list)
	}
}

func TestSubmissionRepositoriesFilterUpdateAndReplaceResults(t *testing.T) {
	db := openRepositoryTestDB(t)
	submissionRepo := NewSubmissionRepository(db)
	resultRepo := NewSubmissionResultRepository(db)
	contestID := uint64(3)
	now := time.Now().UTC()
	runtime := 15

	submissions := []model.Submission{
		{UserID: 1, ProblemID: 10, Language: "cpp", Status: "Accepted", RuntimeMS: &runtime, CreatedAt: now.Add(time.Minute)},
		{UserID: 1, ProblemID: 11, ContestID: &contestID, Language: "go", Status: "Wrong Answer", CreatedAt: now.Add(2 * time.Minute)},
		{UserID: 2, ProblemID: 10, Language: "cpp", Status: "Pending", CreatedAt: now},
	}
	for i := range submissions {
		if err := submissionRepo.Create(&submissions[i]); err != nil {
			t.Fatalf("create submission %d: %v", i, err)
		}
	}

	status := "Wrong Answer"
	filtered, total, err := submissionRepo.ListByUserID(1, 1, 10, nil, &contestID, &status)
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ProblemID != 11 {
		t.Fatalf("unexpected filtered submissions total=%d list=%+v", total, filtered)
	}

	recent, err := submissionRepo.ListRecentByUserID(1, 1, 10, nil, nil, nil)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(recent) != 2 || recent[0].ProblemID != 11 {
		t.Fatalf("unexpected recent submissions: %+v", recent)
	}

	first, err := submissionRepo.GetByID(submissions[0].ID)
	if err != nil {
		t.Fatalf("get submission: %v", err)
	}
	first.Status = "Rejudged"
	if err := submissionRepo.Update(db, first); err != nil {
		t.Fatalf("update submission: %v", err)
	}
	updated, err := submissionRepo.GetByID(first.ID)
	if err != nil {
		t.Fatalf("get updated submission: %v", err)
	}
	if updated.Status != "Rejudged" {
		t.Fatalf("expected updated status, got %q", updated.Status)
	}

	results := []model.SubmissionResult{
		{SubmissionID: first.ID, TestcaseID: 2, Status: "Wrong Answer"},
		{SubmissionID: first.ID, TestcaseID: 1, Status: "Accepted"},
	}
	if err := resultRepo.ReplaceBySubmissionID(db, first.ID, results); err != nil {
		t.Fatalf("replace results: %v", err)
	}
	list, err := resultRepo.ListBySubmissionID(first.ID)
	if err != nil {
		t.Fatalf("list results: %v", err)
	}
	if len(list) != 2 || list[0].TestcaseID != 2 || list[1].TestcaseID != 1 {
		t.Fatalf("unexpected result order: %+v", list)
	}

	if err := resultRepo.ReplaceBySubmissionID(db, first.ID, nil); err != nil {
		t.Fatalf("clear results: %v", err)
	}
	list, err = resultRepo.ListBySubmissionID(first.ID)
	if err != nil {
		t.Fatalf("list cleared results: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected cleared results, got %+v", list)
	}
}
