package service

import (
	"errors"
	"testing"
	"time"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/repository"
)

func TestContestServicePrivateAndUpcomingAccess(t *testing.T) {
	db := openServiceTestDB(t)
	problemRepo := repository.NewProblemRepository(db)
	testcaseRepo := repository.NewProblemTestcaseRepository(db)
	svc := NewContestService(db, problemRepo, testcaseRepo, nil)

	problem := model.Problem{
		DisplayID: "CT-PRIV", Title: "P", Description: "d", JudgeMode: "standard",
		TimeLimitMS: 1000, MemoryLimitMB: 128, Visible: true, CreatedBy: 1,
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	user := model.User{Username: "ctuser", UserID: "CTU01", PasswordHash: "h", Role: "student", Status: "active"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Now()
	privateID, err := svc.Create(1, "admin", dto.CreateContestRequest{
		Title: "Private", RuleType: "acm",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour),
		IsPublic: false, Problems: []dto.CreateContestProblemDTO{{ProblemID: problem.ID}},
	})
	if err != nil {
		t.Fatalf("create private contest: %v", err)
	}
	if err := db.Model(&model.Contest{}).Where("id = ?", privateID).Update("is_public", false).Error; err != nil {
		t.Fatalf("force private contest: %v", err)
	}

	if _, err := svc.ListProblems(user.ID, "student", privateID); !errors.Is(err, ErrContestForbidden) {
		t.Fatalf("expected private contest forbidden, got %v", err)
	}

	upcomingID, err := svc.Create(1, "admin", dto.CreateContestRequest{
		Title: "Upcoming", RuleType: "acm",
		StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour),
		IsPublic: true, Problems: []dto.CreateContestProblemDTO{{ProblemID: problem.ID}},
	})
	if err != nil {
		t.Fatalf("create upcoming contest: %v", err)
	}
	if _, err := svc.ListProblems(user.ID, "student", upcomingID); !errors.Is(err, ErrContestForbidden) {
		t.Fatalf("expected upcoming contest forbidden, got %v", err)
	}

	if _, err := svc.GetProblemDetail(1, "admin", privateID, 99999); !errors.Is(err, ErrContestProblemNotFound) {
		t.Fatalf("expected contest problem not found, got %v", err)
	}
}

func TestProblemServiceExportNotFound(t *testing.T) {
	db := openServiceTestDB(t)
	problemRepo := repository.NewProblemRepository(db)
	testcaseRepo := repository.NewProblemTestcaseRepository(db)
	svc := NewProblemService(db, problemRepo, testcaseRepo, nil)

	if _, _, err := svc.ExportProblemPackage(99999); !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected export package not found, got %v", err)
	}
	if _, _, err := svc.ExportProblemTestcases(99999); !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected export testcases not found, got %v", err)
	}
}
