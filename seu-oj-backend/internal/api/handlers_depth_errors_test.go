package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/model"
)

func callHandlerUserOnly(t *testing.T, method, path string, body any, params gin.Params, userID uint64, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, reader)
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	c.Params = params
	if userID > 0 {
		c.Set(middleware.ContextUserIDKey, userID)
	}
	handler(c)
	return w
}

func seedProblem(t *testing.T, env *apiTestEnv, displayID string, visible bool) uint64 {
	t.Helper()
	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": displayID, "title": displayID, "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": visible,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	decodeOKData(t, w, &problemID)
	return problemID.ProblemID
}

func TestAPIContestAccessDepthErrors(t *testing.T) {
	env := newAPITestEnv(t)
	problemID := seedProblem(t, env, "DEPTH-P1", true)
	now := time.Now()

	w := callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Private", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": false, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": problemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var privateContest struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &privateContest)

	w = callHandler(t, http.MethodGet, "/contests/:id/problems", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}}, env.studentID, "student", env.contest.Problems)
	if decodeAPIEnvelope(t, w).Message != "contest problems unavailable" {
		t.Fatalf("expected private contest blocked, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Upcoming", "description": "d", "rule_type": "acm",
		"start_time": now.Add(time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(2 * time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": problemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var upcomingContest struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &upcomingContest)
	w = callHandler(t, http.MethodGet, "/contests/:id/problems", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(upcomingContest.ContestID)}}, env.studentID, "student", env.contest.Problems)
	if decodeAPIEnvelope(t, w).Message != "contest problems unavailable" {
		t.Fatalf("expected upcoming contest blocked, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandlerUserOnly(t, http.MethodGet, "/contests/:id/problems", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(upcomingContest.ContestID)}}, env.studentID, env.contest.Problems)
	if decodeAPIEnvelope(t, w).Message != "missing user role" {
		t.Fatalf("expected missing role, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodGet, "/admin/contests/:id/ranklist", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}}, env.studentID, "student", env.contest.AdminRanklist)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatal("expected admin ranklist denied")
	}

	w = callHandler(t, http.MethodGet, "/admin/contests/:id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}}, env.studentID, "student", env.contest.AdminDetail)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatal("expected admin detail denied")
	}

	w = callHandler(t, http.MethodPut, "/admin/contests/:id", map[string]any{
		"title": "Bad Time", "description": "d", "rule_type": "acm",
		"start_time": now.Add(time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(-time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": problemID}},
	}, gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}}, env.studentID, "student", env.contest.Update)
	if decodeAPIEnvelope(t, w).Message != "permission denied" {
		t.Fatal("expected student update denied")
	}

	w = callHandler(t, http.MethodPut, "/admin/contests/:id", map[string]any{
		"title": "Bad Time", "description": "d", "rule_type": "acm",
		"start_time": now.Add(time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(-time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": false,
		"problems": []map[string]any{{"problem_id": problemID}},
	}, gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}}, env.adminID, "admin", env.contest.Update)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid time range")
	}

	w = callHandler(t, http.MethodGet, "/contests/:id/problems/:problem_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(privateContest.ContestID)}, {Key: "problem_id", Value: "99999"}},
		env.adminID, "admin", env.contest.ProblemDetail)
	if decodeAPIEnvelope(t, w).Message != "contest problem not found" {
		t.Fatalf("expected contest problem not found, got %q", decodeAPIEnvelope(t, w).Message)
	}
}

func TestAPIContestAnnouncementDepthErrors(t *testing.T) {
	env := newAPITestEnv(t)
	problemID := seedProblem(t, env, "DEPTH-P2", true)
	now := time.Now()

	w := callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "Ann Contest", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": true,
		"problems": []map[string]any{{"problem_id": problemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	var contestID struct {
		ContestID uint64 `json:"contest_id"`
	}
	decodeOKData(t, w, &contestID)

	w = callHandler(t, http.MethodGet, "/contests/:id/announcements?page=bad", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, 0, "", env.contest.AnnouncementList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid announcement list query")
	}

	w = callHandler(t, http.MethodGet, "/contests/:id/announcements/:announcement_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: "99999"}},
		0, "", env.contest.AnnouncementDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected missing contest announcement")
	}

	w = callHandler(t, http.MethodPost, "/admin/contests/:id/announcements", map[string]any{"title": ""},
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.studentID, "student", env.contest.CreateAnnouncement)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected create contest announcement denied")
	}

	w = callHandler(t, http.MethodPut, "/admin/contests/:id/announcements/:announcement_id", map[string]any{
		"title": "x", "content": "y",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: "99999"}},
		env.adminID, "admin", env.contest.UpdateAnnouncement)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected update missing contest announcement")
	}

	w = callHandler(t, http.MethodDelete, "/admin/contests/:id/announcements/:announcement_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: "99999"}},
		env.adminID, "admin", env.contest.DeleteAnnouncement)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected delete missing contest announcement")
	}
}

func TestAPIProblemDepthErrors(t *testing.T) {
	env := newAPITestEnv(t)
	hiddenID := seedProblem(t, env, "DEPTH-HID", false)
	visibleID := seedProblem(t, env, "DEPTH-VIS", true)

	w := callHandler(t, http.MethodGet, "/public/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(hiddenID)}}, 0, "", env.problem.PublicDetail)
	if decodeAPIEnvelope(t, w).Message != "problem not found" {
		t.Fatal("expected hidden public detail not found")
	}

	w = callHandler(t, http.MethodGet, "/problems/:id/stats", nil, gin.Params{{Key: "id", Value: fmt.Sprint(99999)}}, 0, "", env.problem.Stats)
	if decodeAPIEnvelope(t, w).Message != "problem not found" {
		t.Fatal("expected stats not found")
	}

	w = callHandler(t, http.MethodGet, "/problems/:id/solutions", nil, gin.Params{{Key: "id", Value: fmt.Sprint(99999)}}, 0, "", env.problem.SolutionList)
	if decodeAPIEnvelope(t, w).Message != "problem not found" {
		t.Fatal("expected solution list not found")
	}

	w = callHandler(t, http.MethodPost, "/teacher/problems/:id/solutions", map[string]any{
		"title": "No AC", "content": "hint", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(visibleID)}}, env.studentID, "student", env.problem.CreateSolution)
	if decodeAPIEnvelope(t, w).Message != "accepted submission required before publishing a solution" {
		t.Fatalf("expected solution publish blocked, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodPut, "/teacher/problems/:id/solutions/:solution_id", map[string]any{
		"title": "x", "content": "y", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(visibleID)}, {Key: "solution_id", Value: "99999"}},
		env.adminID, "admin", env.problem.UpdateSolution)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected update missing solution error")
	}

	w = callHandler(t, http.MethodDelete, "/teacher/problems/:id/solutions/:solution_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(visibleID)}, {Key: "solution_id", Value: "99999"}},
		env.adminID, "admin", env.problem.DeleteSolution)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected delete missing solution error")
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(99999)}}, env.adminID, "admin", env.problem.AdminDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected admin detail not found")
	}
}

func TestAPIForumAndAuthDepthErrors(t *testing.T) {
	env := newAPITestEnv(t)

	disabledHash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	disabled := model.User{Username: "disabledu", UserID: "DIS001", PasswordHash: string(disabledHash), Role: "student", Status: "disabled"}
	if err := env.db.Create(&disabled).Error; err != nil {
		t.Fatalf("seed disabled user: %v", err)
	}

	w := callHandler(t, http.MethodPost, "/login", map[string]any{
		"username": "disabledu", "password": "secret1",
	}, nil, 0, "", env.auth.Login)
	if decodeAPIEnvelope(t, w).Message != "user is disabled" {
		t.Fatalf("expected disabled login failure, got %q", decodeAPIEnvelope(t, w).Message)
	}

	w = callHandler(t, http.MethodGet, "/me", nil, nil, 99999, "student", env.auth.Me)
	if decodeAPIEnvelope(t, w).Message != "user not found" {
		t.Fatal("expected me not found")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics/:id/like", nil, gin.Params{{Key: "id", Value: "bad"}}, env.studentID, "student", env.forum.LikeTopic)
	if decodeAPIEnvelope(t, w).Message != "invalid topic id" {
		t.Fatal("expected invalid like topic id")
	}

	w = callHandler(t, http.MethodDelete, "/forum/topics/:id/like", nil, gin.Params{{Key: "id", Value: "bad"}}, env.studentID, "student", env.forum.UnlikeTopic)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid unlike topic id")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics/:id/favorite", nil, gin.Params{{Key: "id", Value: "bad"}}, env.studentID, "student", env.forum.FavoriteTopic)
	if decodeAPIEnvelope(t, w).Message != "invalid topic id" {
		t.Fatal("expected invalid favorite topic id")
	}

	w = callHandler(t, http.MethodGet, "/ranklist?page=bad", nil, nil, 0, "", env.ranklist.List)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid ranklist query")
	}

	w = callHandler(t, http.MethodGet, "/submissions/public?page_size=101", nil, nil, 0, "", env.submission.PublicList)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid public submission query")
	}
}

func TestAPITeachingDepthErrors(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandlerUserOnly(t, http.MethodGet, "/teacher/classes/:id", nil,
		gin.Params{{Key: "id", Value: "1"}}, env.teacherID, env.teaching.TeacherClassDetail)
	if decodeAPIEnvelope(t, w).Message != "missing user role" {
		t.Fatal("expected missing role on teacher class detail")
	}

	w = callHandler(t, http.MethodGet, "/teacher/classes/:id/analytics", nil,
		gin.Params{{Key: "id", Value: "99999"}}, env.teacherID, "teacher", env.teaching.TeacherClassAnalytics)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected analytics not found")
	}

	w = callHandler(t, http.MethodPut, "/teacher/classes/:id", map[string]any{
		"name": "x", "description": "d", "status": "active",
	}, gin.Params{{Key: "id", Value: "99999"}}, env.teacherID, "teacher", env.teaching.UpdateClass)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected update missing class error")
	}

	w = callHandler(t, http.MethodGet, "/playlists/:id", nil, gin.Params{{Key: "id", Value: "99999"}}, env.studentID, "student", env.teaching.PlaylistDetail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected playlist not found")
	}
}
