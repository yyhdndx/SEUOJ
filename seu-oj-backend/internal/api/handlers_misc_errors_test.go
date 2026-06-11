package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIAnnouncementHandlerErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/announcements?page=bad", nil, nil, 0, "", env.announcement.List)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid announcement list query")
	}

	w = callHandler(t, http.MethodGet, "/announcements/:id", nil, gin.Params{{Key: "id", Value: "bad"}}, 0, "", env.announcement.Detail)
	if decodeAPIEnvelope(t, w).Message != "invalid announcement id" {
		t.Fatal("expected invalid announcement id")
	}

	w = callHandler(t, http.MethodPost, "/admin/announcements", map[string]any{
		"title": "Denied", "content": "body",
	}, nil, env.studentID, "student", env.announcement.Create)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatalf("expected student create announcement denied, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPut, "/admin/announcements/:id", map[string]any{"title": "x", "content": "body"},
		gin.Params{{Key: "id", Value: "bad"}}, env.adminID, "admin", env.announcement.Update)
	if decodeAPIEnvelope(t, w).Message != "invalid announcement id" {
		t.Fatal("expected invalid update id")
	}

	w = callHandler(t, http.MethodPut, "/admin/announcements/:id", map[string]any{"title": "x", "content": "body"},
		gin.Params{{Key: "id", Value: "99999"}}, env.adminID, "admin", env.announcement.Update)
	if decodeAPIEnvelope(t, w).Message != "announcement not found" {
		t.Fatalf("expected update not found, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPut, "/admin/announcements/:id", map[string]any{"title": "x", "content": "body"},
		gin.Params{{Key: "id", Value: "1"}}, env.studentID, "student", env.announcement.Update)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatal("expected student update denied")
	}

	w = callHandler(t, http.MethodDelete, "/admin/announcements/:id", nil,
		gin.Params{{Key: "id", Value: "bad"}}, env.adminID, "admin", env.announcement.Delete)
	if decodeAPIEnvelope(t, w).Message != "invalid announcement id" {
		t.Fatal("expected invalid delete id")
	}

	w = callHandler(t, http.MethodDelete, "/admin/announcements/:id", nil,
		gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.announcement.Delete)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatal("expected student delete denied")
	}
}

func TestAPIAuthProfileErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPut, "/profile", map[string]any{"username": ""},
		nil, env.studentID, "student", env.auth.UpdateProfile)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected profile validation error")
	}

	w = callHandler(t, http.MethodPut, "/password", map[string]any{"current_password": "x"},
		nil, env.studentID, "student", env.auth.ChangePassword)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected password bind error")
	}
}

func TestAPIContestHandlerParseErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/contests/:id", nil, gin.Params{{Key: "id", Value: "bad"}}, 0, "", env.contest.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid contest id")
	}

	w = callHandler(t, http.MethodGet, "/contests/:id/problems", nil, gin.Params{{Key: "id", Value: "bad"}}, env.studentID, "student", env.contest.Problems)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid contest id for problems")
	}

	w = callHandler(t, http.MethodGet, "/contests/:id/ranklist", nil, gin.Params{{Key: "id", Value: fmt.Sprint(99999)}}, 0, "", env.contest.Ranklist)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected ranklist not found")
	}

	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{"title": ""},
		nil, env.studentID, "student", env.contest.Create)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected contest create permission/bind error")
	}
}

func TestAPIForumHandlerErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/forum/topics?page=bad", nil, nil, 0, "", env.forum.ListTopics)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid forum list query")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{
		"title": "Scoped", "content": "body", "scope_type": "problem",
	}, nil, env.studentID, "student", env.forum.CreateTopic)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected scoped topic without scope id error")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{
		"title": "Missing Problem", "content": "body", "scope_type": "problem", "scope_id": 99999,
	}, nil, env.studentID, "student", env.forum.CreateTopic)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected missing problem scope error")
	}

	w = callHandler(t, http.MethodPut, "/forum/topics/:id", map[string]any{"title": "hack"},
		gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.forum.UpdateTopic)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected update missing topic error")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics/:id/replies", map[string]any{"content": "x"},
		gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.forum.CreateReply)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected reply to missing topic error")
	}
}

func TestAPIProblemHandlerQueryErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/problems?page=bad", nil, nil, 0, "", env.problem.List)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid problem list query")
	}

	w = callHandler(t, http.MethodGet, "/admin/problems?page=bad", nil, nil, env.adminID, "admin", env.problem.AdminList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid admin problem list query")
	}

	w = callHandler(t, http.MethodGet, "/public/problems?page=bad", nil, nil, 0, "", env.problem.PublicList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid public problem list query")
	}

	w = callHandler(t, http.MethodGet, "/teacher/problems/:id/solutions", nil,
		gin.Params{{Key: "id", Value: "bad"}}, env.teacherID, "teacher", env.problem.TeacherSolutionList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid teacher solution list id")
	}
}

func TestAPITeachingHandlerErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/classes/join", map[string]any{"join_code": "NOPE"},
		nil, env.studentID, "student", env.teaching.JoinClass)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid join code error")
	}

	w = callHandler(t, http.MethodGet, "/teacher/classes/:id/members", nil,
		gin.Params{{Key: "id", Value: "bad"}}, env.teacherID, "teacher", env.teaching.ClassMembers)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid class members id")
	}

	w = callHandler(t, http.MethodPost, "/teacher/classes/:id/assignments", map[string]any{"title": ""},
		gin.Params{{Key: "id", Value: "99999"}}, env.teacherID, "teacher", env.teaching.CreateAssignment)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected create assignment error")
	}
}
