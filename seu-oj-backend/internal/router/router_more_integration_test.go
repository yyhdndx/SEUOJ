package router

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"seu-oj-backend/internal/testutil"
)

func migrateRouterTeachingSQLite(t *testing.T, db *gorm.DB) {
	t.Helper()
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			visibility TEXT NOT NULL DEFAULT 'public',
			created_by INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS playlist_problems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			problem_id INTEGER NOT NULL,
			display_order INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			join_code TEXT NOT NULL UNIQUE,
			teacher_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS class_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			class_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			role TEXT NOT NULL DEFAULT 'student',
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS assignments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			class_id INTEGER NOT NULL,
			playlist_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL DEFAULT 'homework',
			start_at DATETIME,
			due_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("migrate teaching table: %v", err)
		}
	}
}

func TestForumAndStatsIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	db := openRouterTestDB(t)
	engine := New(db, redisClient, routerTestConfig())
	studentToken := seedRouterUserAndToken(t, db, "forumuser", "student")

	resp := doJSON(t, engine, http.MethodGet, "/api/forum/topics?page=1&page_size=10", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPost, "/api/forum/topics", studentToken, map[string]any{
		"title":      "Router Topic",
		"content":    "From router test",
		"scope_type": "general",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	topic := decodeData[struct {
		ID uint64 `json:"id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/forum/topics/%d", topic.ID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/stats/overview", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/stats/me", studentToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	adminToken := seedRouterUserAndToken(t, db, "statsadmin", "admin")
	resp = doJSON(t, engine, http.MethodGet, "/api/stats/admin", adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func TestContestPublicIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRouterTestDB(t)
	engine := New(db, nil, routerTestConfig())
	adminToken := seedRouterUserAndToken(t, db, "contestadmin", "admin")

	resp := doJSON(t, engine, http.MethodPost, "/api/admin/problems", adminToken, problemPayload("C-PUB", "Contest Problem", 1, true), http.StatusOK)
	expectAppCode(t, resp, 0)
	created := decodeData[struct {
		ProblemID uint64 `json:"problem_id"`
	}](t, resp)

	now := time.Now()
	resp = doJSON(t, engine, http.MethodPost, "/api/admin/contests", adminToken, map[string]any{
		"title":          "Public Contest",
		"description":    "demo",
		"rule_type":      "acm",
		"start_time":     now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":       now.Add(time.Hour).Format(time.RFC3339),
		"is_public":      true,
		"allow_practice": true,
		"problems":         []map[string]any{{"problem_id": created.ProblemID}},
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	contest := decodeData[struct {
		ContestID uint64 `json:"contest_id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodGet, "/api/public/contests?page=1&page_size=10", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	list := decodeData[struct {
		Total int64 `json:"total"`
	}](t, resp)
	if list.Total != 1 {
		t.Fatalf("expected public contest list, got %+v", list)
	}

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/contests/%d", contest.ContestID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/contests/%d/ranklist", contest.ContestID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/admin/contests?page=1&page_size=10", adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func TestTeachingRouterIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRouterTestDB(t)
	migrateRouterTeachingSQLite(t, db)
	engine := New(db, nil, routerTestConfig())
	teacherToken := seedRouterUserAndToken(t, db, "teacher1", "teacher")
	adminToken := seedRouterUserAndToken(t, db, "tadmin", "admin")

	resp := doJSON(t, engine, http.MethodPost, "/api/admin/problems", adminToken, problemPayload("T-ROUTER", "Teach Problem", 1, true), http.StatusOK)
	expectAppCode(t, resp, 0)
	problem := decodeData[struct {
		ProblemID uint64 `json:"problem_id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodPost, "/api/teacher/playlists", teacherToken, map[string]any{
		"title": "Router Playlist", "description": "demo", "visibility": "public",
		"problems": []map[string]any{{"problem_id": problem.ProblemID, "display_order": 1}},
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	playlist := decodeData[struct {
		ID uint64 `json:"id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodGet, "/api/playlists?page=1&page_size=10", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/playlists/%d", playlist.ID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPost, "/api/teacher/classes", teacherToken, map[string]any{
		"name": "Router Class", "description": "demo",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	classDetail := decodeData[struct {
		ID       uint64 `json:"id"`
		JoinCode string `json:"join_code"`
	}](t, resp)

	studentToken := seedRouterUserAndToken(t, db, "classstudent", "student")
	resp = doJSON(t, engine, http.MethodPost, "/api/classes/join", studentToken, map[string]any{
		"join_code": classDetail.JoinCode,
	}, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/classes/my", studentToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPost, fmt.Sprintf("/api/teacher/classes/%d/assignments", classDetail.ID), teacherToken, map[string]any{
		"playlist_id": playlist.ID, "title": "HW Router", "type": "homework",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	assignment := decodeData[struct {
		ID uint64 `json:"id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/assignments/%d", assignment.ID), studentToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func TestSubmissionRunIntegration(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	testutil.EnsureDockerImage(t, "python:3")

	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	db := openRouterTestDB(t)
	cfg := routerTestConfig()
	cfg.Sandbox.DockerImage = "python:3"
	engine := New(db, redisClient, cfg)
	adminToken := seedRouterUserAndToken(t, db, "runadmin", "admin")
	studentToken := seedRouterUserAndToken(t, db, "runstudent", "student")

	resp := doJSON(t, engine, http.MethodPost, "/api/admin/problems", adminToken, map[string]any{
		"display_id":      "RUN-R",
		"title":           "Run Sample",
		"description":     "add",
		"input_desc":      "in",
		"output_desc":     "out",
		"judge_mode":      "standard",
		"difficulty":      1,
		"time_limit_ms":   2000,
		"memory_limit_mb": 128,
		"visible":         true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1 2\n", "output_data": "3\n", "sort_order": 1, "is_active": true},
		},
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	created := decodeData[struct {
		ProblemID uint64 `json:"problem_id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodPost, "/api/submissions/run", studentToken, map[string]any{
		"problem_id": created.ProblemID,
		"language":   "python3",
		"code":       "a, b = map(int, input().split())\nprint(a + b)\n",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	runResult := decodeData[struct {
		Status string `json:"status"`
	}](t, resp)
	if runResult.Status != "Accepted" {
		t.Fatalf("unexpected run result: %+v", runResult)
	}

	resp = doJSON(t, engine, http.MethodGet, "/api/public/submissions?page=1&page_size=10", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func TestAdminUsersAndAnnouncementsIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	db := openRouterTestDB(t)
	engine := New(db, redisClient, routerTestConfig())
	adminToken := seedRouterUserAndToken(t, db, "admusr", "admin")
	studentToken := seedRouterUserAndToken(t, db, "admstu", "student")

	resp := doJSON(t, engine, http.MethodGet, "/api/admin/users?page=1&page_size=10", adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/admin/users?page=1&page_size=10", studentToken, nil, http.StatusForbidden)

	resp = doJSON(t, engine, http.MethodPost, "/api/admin/announcements", adminToken, map[string]any{
		"title": "Router Ann", "content": "hello", "is_pinned": true,
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	ann := decodeData[struct {
		AnnouncementID uint64 `json:"announcement_id"`
	}](t, resp)

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/announcements/%d", ann.AnnouncementID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPut, fmt.Sprintf("/api/admin/announcements/%d", ann.AnnouncementID), adminToken, map[string]any{
		"title": "Updated", "content": "body", "is_pinned": false,
	}, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/ranklist?page=1&page_size=20", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodDelete, fmt.Sprintf("/api/admin/announcements/%d", ann.AnnouncementID), adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}
