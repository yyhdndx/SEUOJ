package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPISubmissionDetailForbidden(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "FORB-P1", "title": "Forbidden", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &problemID)

	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": problemID.ProblemID, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	var subID struct {
		SubmissionID uint64 `json:"submission_id"`
	}
	decodeOKData(t, w, &subID)

	w = callHandler(t, http.MethodGet, "/submissions/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(subID.SubmissionID)}}, env.teacherID, "teacher", env.submission.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected forbidden submission detail")
	}

	w = callHandler(t, http.MethodGet, "/submissions/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(subID.SubmissionID)}}, env.adminID, "admin", env.submission.Detail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin should read submission detail")
	}
}

func TestAPIProblemCreateValidation(t *testing.T) {
	env := newAPITestEnv(t)
	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "", "title": "", "judge_mode": "bad",
	}, nil, env.adminID, "admin", env.problem.Create)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected create validation error")
	}
}
