package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIHandlerErrorMatrix(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/admin/problems/import", nil, nil, env.studentID, "student", env.problem.ImportProblemPackage)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected import permission/missing file error")
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id/export", nil, gin.Params{{Key: "id", Value: "99999"}}, env.adminID, "admin", env.problem.ExportProblemPackage)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected export not found")
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id/testcases/export", nil, gin.Params{{Key: "id", Value: "99999"}}, env.adminID, "admin", env.problem.ExportTestcases)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected testcase export not found")
	}

	w = callHandler(t, http.MethodDelete, "/admin/problems/:id", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.problem.Delete)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected delete permission denied")
	}

	w = callHandler(t, http.MethodGet, "/admin/submissions?page=1&page_size=10", nil, nil, env.studentID, "student", env.submission.AdminList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected admin submission list denied")
	}

	w = callHandler(t, http.MethodPost, "/admin/submissions/:id/rejudge", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.submission.Rejudge)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected rejudge permission denied")
	}

	w = callHandler(t, http.MethodGet, "/stats/admin", nil, nil, env.studentID, "student", env.stats.Admin)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected stats admin denied")
	}

	w = callHandler(t, http.MethodPut, "/admin/users/:id", map[string]any{
		"username": "nope", "userid": "NOPE", "role": "admin", "status": "active",
	}, gin.Params{{Key: "id", Value: "bad"}}, env.adminID, "admin", env.user.AdminUpdate)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid user id")
	}

	w = callHandler(t, http.MethodGet, "/contests/:id/me", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.contest.Me)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected contest me not found")
	}

	w = callHandler(t, http.MethodPost, "/contests/:id/register", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.contest.Register)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected contest register not found")
	}

	w = callHandler(t, http.MethodGet, "/teacher/playlists/:id", nil, gin.Params{{Key: "id", Value: "99999"}}, env.teacherID, "teacher", env.teaching.TeacherPlaylistDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected playlist not found")
	}

	w = callHandler(t, http.MethodGet, "/assignments/:id", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.teaching.AssignmentDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected assignment not found")
	}

	w = callHandler(t, http.MethodGet, "/classes/:id", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.teaching.ClassDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected class not found")
	}

	w = callHandler(t, http.MethodPost, "/classes/join", map[string]any{"join_code": "INVALID"},
		nil, env.studentID, "student", env.teaching.JoinClass)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid join code")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics/:id/replies", map[string]any{"content": ""},
		gin.Params{{Key: "id", Value: fmt.Sprint(1)}}, env.studentID, "student", env.forum.CreateReply)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected empty reply error")
	}
}
