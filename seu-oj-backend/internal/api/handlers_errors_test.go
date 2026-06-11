package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAPIHandlersErrorPaths(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/register", map[string]any{"username": "x"}, nil, 0, "", env.auth.Register)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected register bind error")
	}
	w = callHandler(t, http.MethodGet, "/me", nil, nil, 0, "", env.auth.Me)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected me without auth error")
	}

	w = callHandler(t, http.MethodGet, "/admin/users", nil, nil, env.studentID, "student", env.user.AdminList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected student admin list denied")
	}

	w = callHandler(t, http.MethodGet, "/submissions/my?page_size=101", nil, nil, env.studentID, "student", env.submission.ListMy)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected list my invalid query")
	}

	w = callHandler(t, http.MethodGet, "/forum/topics/:id", nil, gin.Params{{Key: "id", Value: "bad"}}, 0, "", env.forum.TopicDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid topic id")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{"title": ""}, nil, env.studentID, "student", env.forum.CreateTopic)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected create topic validation error")
	}
}

func TestAPIHandlersLifecycleExtended(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "EXT-P1", "title": "Ext", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &problemID)

	now := time.Now()
	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Ext Contest", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": true,
		"problems": []map[string]any{{"problem_id": problemID.ProblemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var contestID struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &contestID)

	w = callHandler(t, http.MethodPut, "/admin/contests/:id", map[string]any{
		"title": "Ext Updated", "description": "d2", "rule_type": "acm",
		"start_time": now.Add(-2 * time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(2 * time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": true,
		"problems": []map[string]any{{"problem_id": problemID.ProblemID}},
	}, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.Update)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update contest failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/playlists", map[string]any{
		"title": "Ext PL", "description": "d", "visibility": "public",
		"problems": []map[string]any{{"problem_id": problemID.ProblemID, "display_order": 1}},
	}, nil, env.teacherID, "teacher", env.teaching.CreatePlaylist)
	var playlistID struct {
		ID uint64 `json:"id"`
	}
	decodeOKData(t, w, &playlistID)

	w = callHandler(t, http.MethodPut, "/teacher/playlists/:id", map[string]any{
		"title": "Ext PL v2", "description": "d2", "visibility": "public",
		"problems": []map[string]any{{"problem_id": problemID.ProblemID, "display_order": 1}},
	}, gin.Params{{Key: "id", Value: fmt.Sprint(playlistID.ID)}}, env.teacherID, "teacher", env.teaching.UpdatePlaylist)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update playlist failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/playlists/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(playlistID.ID)}}, env.teacherID, "teacher", env.teaching.TeacherPlaylistDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher playlist detail failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/classes", map[string]any{
		"name": "Ext Class", "description": "d",
	}, nil, env.teacherID, "teacher", env.teaching.CreateClass)
	var classInfo struct {
		ID       uint64 `json:"id"`
		JoinCode string `json:"join_code"`
	}
	decodeOKData(t, w, &classInfo)
	w = callHandler(t, http.MethodPost, "/classes/join", map[string]any{"join_code": classInfo.JoinCode},
		nil, env.studentID, "student", env.teaching.JoinClass)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("join class failed")
	}
	w = callHandler(t, http.MethodGet, "/classes/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.studentID, "student", env.teaching.ClassDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("class detail failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/classes/:id/assignments", map[string]any{
		"playlist_id": playlistID.ID, "title": "HW-Ext", "description": "d", "type": "homework",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.CreateAssignment)
	var assignmentID struct {
		ID uint64 `json:"id"`
	}
	decodeOKData(t, w, &assignmentID)
	w = callHandler(t, http.MethodPut, "/teacher/assignments/:id", map[string]any{
		"playlist_id": playlistID.ID, "title": "HW-Ext v2", "description": "d2", "type": "exam",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(assignmentID.ID)}}, env.teacherID, "teacher", env.teaching.UpdateAssignment)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update assignment failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/assignments/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(assignmentID.ID)}}, env.teacherID, "teacher", env.teaching.TeacherAssignmentOverview)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("assignment overview failed")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{
		"title": "Delete Me", "content": "body", "scope_type": "general",
	}, nil, env.studentID, "student", env.forum.CreateTopic)
	var topicID struct {
		ID uint64 `json:"id"`
	}
	decodeOKData(t, w, &topicID)
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/replies", map[string]any{"content": "edit me"},
		gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.CreateReply)
	var replyID struct {
		ID uint64 `json:"id"`
	}
	decodeOKData(t, w, &replyID)
	w = callHandler(t, http.MethodPut, "/forum/topics/:id/replies/:reply_id", map[string]any{"content": "edited"},
		gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}, {Key: "reply_id", Value: fmt.Sprint(replyID.ID)}},
		env.studentID, "student", env.forum.UpdateReply)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update reply failed")
	}
	w = callHandler(t, http.MethodDelete, "/forum/topics/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.DeleteTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete topic failed")
	}

	w = callHandler(t, http.MethodDelete, "/teacher/assignments/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(assignmentID.ID)}}, env.teacherID, "teacher", env.teaching.DeleteAssignment)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete assignment failed")
	}
	w = callHandler(t, http.MethodDelete, "/teacher/playlists/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(playlistID.ID)}}, env.teacherID, "teacher", env.teaching.DeletePlaylist)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete playlist failed")
	}
	w = callHandler(t, http.MethodDelete, "/admin/contests/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.Delete)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete contest failed")
	}
}

func decodeOKData(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	env := decodeAPIEnvelope(t, w)
	if env.Code != 0 {
		t.Fatalf("unexpected api error: %+v body=%s", env, w.Body.String())
	}
	if err := json.Unmarshal(env.Data, v); err != nil {
		t.Fatalf("decode data: %v", err)
	}
}
