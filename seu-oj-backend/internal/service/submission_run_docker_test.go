package service

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/sandbox"
	"seu-oj-backend/internal/testutil"
	"seu-oj-backend/internal/dto"
)

func newDockerSubmissionService(t *testing.T) *SubmissionService {
	t.Helper()
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")

	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	svc.sandboxRunner = sandbox.NewRunner(sandbox.Config{
		Image:        "python:3",
		CompileImage: "python:3",
		RunImage:     "python:3",
	})
	return svc
}

func seedSampleProblem(t *testing.T, svc *SubmissionService, input, output string) uint64 {
	t.Helper()

	problem := model.Problem{
		DisplayID:     "RUN-1",
		Title:         "Run Sample",
		Description:   "desc",
		JudgeMode:     "standard",
		TimeLimitMS:   2000,
		MemoryLimitMB: 128,
		Visible:       true,
		CreatedBy:     1,
	}
	if err := svc.problemRepo.Create(svc.db, &problem); err != nil {
		t.Fatalf("create problem: %v", err)
	}
	testcase := model.ProblemTestcase{
		ProblemID:  problem.ID,
		CaseType:   "sample",
		InputData:  input,
		OutputData: output,
		SortOrder:  1,
		IsActive:   true,
		Score:      0,
	}
	if err := svc.db.Create(&testcase).Error; err != nil {
		t.Fatalf("create sample testcase: %v", err)
	}
	return problem.ID
}

func TestRunSampleTestsAcceptedWithDocker(t *testing.T) {
	svc := newDockerSubmissionService(t)
	problemID := seedSampleProblem(t, svc, "3 4\n", "7\n")
	code := "a, b = map(int, input().split())\nprint(a + b)\n"

	resp, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problemID,
		Language:  "python3",
		Code:      code,
	})
	if err != nil {
		t.Fatalf("run sample tests: %v", err)
	}
	if resp.Status != "Accepted" || len(resp.Results) != 1 || resp.Results[0].Status != "Accepted" {
		t.Fatalf("unexpected run response: %+v", resp)
	}
}

func TestRunSampleTestsWrongAnswerWithDocker(t *testing.T) {
	svc := newDockerSubmissionService(t)
	problemID := seedSampleProblem(t, svc, "2\n", "2\n")

	resp, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problemID,
		Language:  "python3",
		Code:      "print(9)\n",
	})
	if err != nil {
		t.Fatalf("run sample tests: %v", err)
	}
	if resp.Status != "Wrong Answer" {
		t.Fatalf("expected Wrong Answer, got %+v", resp)
	}
}

func TestRunSampleTestsCompileErrorWithDocker(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "gcc:13")

	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	svc.sandboxRunner = sandbox.NewRunner(sandbox.Config{Image: "gcc:13"})

	problemID := seedSampleProblem(t, svc, "1\n", "1\n")
	resp, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problemID,
		Language:  "cpp",
		Code:      "int main() { return nope; }",
	})
	if err != nil {
		t.Fatalf("run sample tests: %v", err)
	}
	if resp.Status != "Compile Error" {
		t.Fatalf("expected Compile Error, got %+v", resp)
	}
}

func TestRunSampleTestsNoSampleCases(t *testing.T) {
	mr := miniredis.RunT(t)
	svc := newTestSubmissionService(t, redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	problemID := seedVisibleProblem(t, svc, "RUN-NOSAMPLE")
	testcase := model.ProblemTestcase{
		ProblemID:  problemID,
		CaseType:   "hidden",
		InputData:  "1",
		OutputData: "1",
		SortOrder:  1,
		IsActive:   true,
		Score:      100,
	}
	if err := svc.db.Create(&testcase).Error; err != nil {
		t.Fatalf("create hidden testcase: %v", err)
	}

	resp, err := svc.RunSampleTests(1, "student", dto.RunSubmissionRequest{
		ProblemID: problemID,
		Language:  "python3",
		Code:      "print(1)\n",
	})
	if err != nil {
		t.Fatalf("run sample tests: %v", err)
	}
	if resp.Status != "System Error" || resp.ErrorMsg != "no sample testcases" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
