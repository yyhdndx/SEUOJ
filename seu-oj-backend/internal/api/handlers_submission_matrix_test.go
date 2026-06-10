package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAPISubmissionCreateErrorMatrix(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/submissions", map[string]any{}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected bind error")
	}

	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": 99999, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if msg := decodeAPIEnvelope(t, w).Message; msg != "problem not found or not visible" {
		t.Fatalf("expected hidden/not found error, got %q", msg)
	}

	w = callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "SUB-HID", "title": "Hidden", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": false,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var hiddenProblem struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &hiddenProblem)
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": hiddenProblem.ProblemID, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Message != "problem not found or not visible" {
		t.Fatal("expected hidden problem rejection")
	}

	w = callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "SUB-VIS", "title": "Visible", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var visibleProblem struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &visibleProblem)

	missingContest := uint64(99999)
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": visibleProblem.ProblemID, "contest_id": missingContest, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Message != "contest not found" {
		t.Fatalf("expected contest not found, got %q", decodeAPIEnvelope(t, w).Message)
	}

	now := time.Now()
	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Future", "description": "d", "rule_type": "acm",
		"start_time": now.Add(time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(2 * time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": visibleProblem.ProblemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var futureContest struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &futureContest)
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": visibleProblem.ProblemID, "contest_id": futureContest.ContestID,
		"language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Message != "contest is not running" {
		t.Fatalf("expected contest not running, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Running", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": visibleProblem.ProblemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var runningContest struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &runningContest)
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": visibleProblem.ProblemID, "contest_id": runningContest.ContestID,
		"language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Message != "contest registration required" {
		t.Fatalf("expected registration required, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "SUB-OTHER", "title": "Other", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var otherProblem struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &otherProblem)
	w = callHandler(t, http.MethodPost, "/contests/:id/register", nil, gin.Params{{Key: "id", Value: fmt.Sprint(runningContest.ContestID)}}, env.studentID, "student", env.contest.Register)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("register contest failed")
	}
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": otherProblem.ProblemID, "contest_id": runningContest.ContestID,
		"language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	if decodeAPIEnvelope(t, w).Message != "problem does not belong to contest" {
		t.Fatalf("expected contest problem mismatch, got %q", decodeAPIEnvelope(t, w).Message)
	}
}

func TestAPISubmissionCreateWithoutAuthContext(t *testing.T) {
	env := newAPITestEnv(t)
	w := callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": 1, "language": "cpp", "code": "int main(){}",
	}, nil, 0, "", env.submission.Create)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected missing user id error")
	}
}

func TestAPISubmissionRunValidation(t *testing.T) {
	env := newAPITestEnv(t)
	w := callHandler(t, http.MethodPost, "/submissions/run", map[string]any{}, nil, env.studentID, "student", env.submission.Run)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected run bind error")
	}
}
