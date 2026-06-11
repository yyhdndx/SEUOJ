package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
	"seu-oj-backend/internal/service"
)

type apiTestEnv struct {
	db           *gorm.DB
	auth         *AuthHandler
	user         *UserHandler
	announcement *AnnouncementHandler
	ranklist     *RanklistHandler
	stats        *StatsHandler
	problem      *ProblemHandler
	submission   *SubmissionHandler
	forum        *ForumHandler
	contest      *ContestHandler
	teaching     *TeachingHandler
	studentID    uint64
	adminID      uint64
	teacherID    uint64
}

func openAPITestDB(t *testing.T) *gorm.DB {
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
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(
		&model.User{}, &model.Problem{}, &model.ProblemTestcase{}, &model.ProblemSolution{},
		&model.Submission{}, &model.SubmissionResult{}, &model.Announcement{},
		&model.Contest{}, &model.ContestProblem{}, &model.ContestRegistration{}, &model.ContestAnnouncement{},
		&model.ForumTopic{}, &model.ForumReply{}, &model.ForumTopicLike{}, &model.ForumTopicFavorite{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	migrateAPITeachingSQLite(t, db)
	return db
}

func migrateAPITeachingSQLite(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, description TEXT,
			visibility TEXT NOT NULL DEFAULT 'public', created_by INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE IF NOT EXISTS playlist_problems (
			id INTEGER PRIMARY KEY AUTOINCREMENT, playlist_id INTEGER NOT NULL,
			problem_id INTEGER NOT NULL, display_order INTEGER NOT NULL DEFAULT 1, created_at DATETIME)`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, description TEXT,
			join_code TEXT NOT NULL UNIQUE, teacher_id INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE IF NOT EXISTS class_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT, class_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
			role TEXT NOT NULL DEFAULT 'student', status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE IF NOT EXISTS assignments (
			id INTEGER PRIMARY KEY AUTOINCREMENT, class_id INTEGER NOT NULL, playlist_id INTEGER NOT NULL,
			title TEXT NOT NULL, description TEXT, type TEXT NOT NULL DEFAULT 'homework',
			start_at DATETIME, due_at DATETIME, created_by INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("migrate teaching: %v", err)
		}
	}
}

func newAPITestEnv(t *testing.T) *apiTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	db := openAPITestDB(t)
	cacheStore := cache.New(redisClient)
	judgeQueue := queue.NewJudgeQueue(redisClient)
	problemRepo := repository.NewProblemRepository(db)
	testcaseRepo := repository.NewProblemTestcaseRepository(db)
	contestSvc := service.NewContestService(db, problemRepo, testcaseRepo, cacheStore)
	env := &apiTestEnv{
		db: db,
		auth:         NewAuthHandler(service.NewAuthService(db, "api-integration-secret")),
		user:         NewUserHandler(service.NewUserService(db)),
		announcement: NewAnnouncementHandler(service.NewAnnouncementService(db, cacheStore)),
		ranklist:     NewRanklistHandler(service.NewRanklistService(db, cacheStore)),
		stats:        NewStatsHandler(service.NewStatsService(db, judgeQueue, cacheStore)),
		problem:      NewProblemHandler(service.NewProblemService(db, problemRepo, testcaseRepo, cacheStore)),
		submission: NewSubmissionHandler(service.NewSubmissionService(db,
			repository.NewSubmissionRepository(db), repository.NewSubmissionResultRepository(db),
			problemRepo, testcaseRepo, judgeQueue, sandbox.NewRunner(sandbox.Config{Image: "gcc:13"}), contestSvc, cacheStore)),
		forum:    NewForumHandler(service.NewForumService(db, cacheStore)),
		contest:  NewContestHandler(contestSvc),
		teaching: NewTeachingHandler(service.NewTeachingService(db)),
	}

	admin := model.User{Username: "apiadmin", UserID: "APIADM", PasswordHash: "h", Role: "admin", Status: "active"}
	teacher := model.User{Username: "apiteacher", UserID: "APITCR", PasswordHash: "h", Role: "teacher", Status: "active"}
	student := model.User{Username: "apistudent", UserID: "APISTU", PasswordHash: "h", Role: "student", Status: "active"}
	if err := db.Create(&[]model.User{admin, teacher, student}).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	var users []model.User
	if err := db.Order("id ASC").Find(&users).Error; err != nil {
		t.Fatalf("load users: %v", err)
	}
	env.adminID, env.teacherID, env.studentID = users[0].ID, users[1].ID, users[2].ID
	return env
}

func callHandler(t *testing.T, method, path string, body any, params gin.Params, userID uint64, role string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
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
		c.Set(middleware.ContextRoleKey, role)
	}
	handler(c)
	return w
}

func decodeAPIEnvelope(t *testing.T, w *httptest.ResponseRecorder) responseBody {
	t.Helper()
	var env responseBody
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, w.Body.String())
	}
	return env
}

type responseBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestAPIHandlersIntegration(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/register", map[string]any{
		"username": "newbie", "userid": "NEW001", "password": "password",
	}, nil, 0, "", env.auth.Register)
	if env := decodeAPIEnvelope(t, w); env.Code != 0 {
		t.Fatalf("register failed: %+v", env)
	}

	w = callHandler(t, http.MethodPost, "/login", map[string]any{
		"username": "apistudent", "password": "wrong",
	}, nil, 0, "", env.auth.Login)
	if env := decodeAPIEnvelope(t, w); env.Code == 0 {
		t.Fatal("expected login failure")
	}

	w = callHandler(t, http.MethodGet, "/me", nil, nil, env.studentID, "student", env.auth.Me)
	if env := decodeAPIEnvelope(t, w); env.Code != 0 {
		t.Fatalf("me failed: %+v", env)
	}

	w = callHandler(t, http.MethodPut, "/profile", map[string]any{
		"username": "apistudent2", "userid": "APIST2",
	}, nil, env.studentID, "student", env.auth.UpdateProfile)
	if env := decodeAPIEnvelope(t, w); env.Code != 0 {
		t.Fatalf("update profile: %+v", env)
	}

	w = callHandler(t, http.MethodPut, "/password", map[string]any{
		"current_password": "old", "new_password": "newpass",
	}, nil, env.studentID, "student", env.auth.ChangePassword)
	if env := decodeAPIEnvelope(t, w); env.Code == 0 {
		t.Fatal("expected password change failure")
	}

	w = callHandler(t, http.MethodPost, "/admin/announcements", map[string]any{
		"title": "Hello", "content": "World", "is_pinned": true,
	}, nil, env.adminID, "admin", env.announcement.Create)
	ann := decodeAPIEnvelope(t, w)
	if ann.Code != 0 {
		t.Fatalf("create announcement: %+v", ann)
	}
	var annID struct {
		AnnouncementID uint64 `json:"announcement_id"`
	}
	if err := json.Unmarshal(ann.Data, &annID); err != nil {
		t.Fatalf("decode ann id: %v", err)
	}

	w = callHandler(t, http.MethodGet, "/announcements?page=1&page_size=10", nil, nil, 0, "", env.announcement.List)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("list announcements failed")
	}
	w = callHandler(t, http.MethodGet, "/announcements/:id", nil, gin.Params{{Key: "id", Value: "999"}}, 0, "", env.announcement.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected missing announcement error")
	}
	w = callHandler(t, http.MethodGet, "/announcements/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(annID.AnnouncementID)}}, 0, "", env.announcement.Detail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("announcement detail failed")
	}
	w = callHandler(t, http.MethodPut, "/admin/announcements/:id", map[string]any{
		"title": "Updated", "content": "Body", "is_pinned": false,
	}, gin.Params{{Key: "id", Value: fmt.Sprint(annID.AnnouncementID)}}, env.adminID, "admin", env.announcement.Update)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update announcement failed")
	}

	w = callHandler(t, http.MethodGet, "/admin/users?page=1&page_size=10", nil, nil, env.adminID, "admin", env.user.AdminList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin user list failed")
	}
	w = callHandler(t, http.MethodPut, "/admin/users/:id", map[string]any{
		"username": "apistudent", "userid": "APISTU", "role": "student", "status": "active",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(env.studentID)}}, env.adminID, "admin", env.user.AdminUpdate)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin user update failed")
	}

	w = callHandler(t, http.MethodGet, "/ranklist?page=1&page_size=20", nil, nil, 0, "", env.ranklist.List)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("ranklist failed")
	}
	w = callHandler(t, http.MethodGet, "/stats/overview", nil, nil, 0, "", env.stats.Overview)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("stats overview failed")
	}
	w = callHandler(t, http.MethodGet, "/stats/me", nil, nil, env.studentID, "student", env.stats.Me)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("stats me failed")
	}
	w = callHandler(t, http.MethodGet, "/stats/admin", nil, nil, env.adminID, "admin", env.stats.Admin)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("stats admin failed")
	}

	w = callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "API-P1", "title": "Add", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1 1", "output_data": "2", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	prob := decodeAPIEnvelope(t, w)
	if prob.Code != 0 {
		t.Fatalf("create problem: %+v", prob)
	}
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	_ = json.Unmarshal(prob.Data, &problemID)

	w = callHandler(t, http.MethodGet, "/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, 0, "", env.problem.Detail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("problem detail failed")
	}
	w = callHandler(t, http.MethodGet, "/public/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, 0, "", env.problem.PublicDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("public problem detail failed")
	}
	w = callHandler(t, http.MethodGet, "/problems/:id/stats", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, 0, "", env.problem.Stats)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("problem stats failed")
	}
	w = callHandler(t, http.MethodPost, "/teacher/problems/:id/solutions", map[string]any{
		"title": "Official", "content": "Use sum", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.CreateSolution)
	sol := decodeAPIEnvelope(t, w)
	if sol.Code != 0 {
		t.Fatalf("create solution: %+v", sol)
	}
	var solutionID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(sol.Data, &solutionID)
	w = callHandler(t, http.MethodGet, "/problems/:id/solutions", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.studentID, "student", env.problem.SolutionList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("solution list failed")
	}
	w = callHandler(t, http.MethodPut, "/teacher/problems/:id/solutions/:solution_id", map[string]any{
		"title": "Official v2", "content": "Updated", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}, {Key: "solution_id", Value: fmt.Sprint(solutionID.ID)}},
		env.adminID, "admin", env.problem.UpdateSolution)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update solution failed")
	}

	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": problemID.ProblemID, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	sub := decodeAPIEnvelope(t, w)
	if sub.Code != 0 {
		t.Fatalf("create submission: %+v", sub)
	}
	var submissionID struct {
		SubmissionID uint64 `json:"submission_id"`
	}
	_ = json.Unmarshal(sub.Data, &submissionID)
	w = callHandler(t, http.MethodGet, "/submissions/my?page=1&page_size=10", nil, nil, env.studentID, "student", env.submission.ListMy)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("list my submissions failed")
	}
	w = callHandler(t, http.MethodGet, "/submissions/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(submissionID.SubmissionID)}}, env.studentID, "student", env.submission.Detail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("submission detail failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/submissions?page=1&page_size=10", nil, nil, env.adminID, "admin", env.submission.AdminList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin submission list failed")
	}
	w = callHandler(t, http.MethodGet, "/public/submissions?page=1&page_size=10", nil, nil, 0, "", env.submission.PublicList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("public submission list failed")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{
		"title": "API Topic", "content": "content", "scope_type": "general",
	}, nil, env.studentID, "student", env.forum.CreateTopic)
	topic := decodeAPIEnvelope(t, w)
	if topic.Code != 0 {
		t.Fatalf("create topic: %+v", topic)
	}
	var topicID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(topic.Data, &topicID)
	w = callHandler(t, http.MethodGet, "/forum/topics?page=1&page_size=10", nil, nil, 0, "", env.forum.ListTopics)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("forum list failed")
	}
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/like", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.LikeTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("like topic failed")
	}
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/replies", map[string]any{"content": "reply"},
		gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.CreateReply)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("create reply failed")
	}

	now := time.Now()
	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "API Contest", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": true,
		"problems": []map[string]any{{"problem_id": problemID.ProblemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	ct := decodeAPIEnvelope(t, w)
	if ct.Code != 0 {
		t.Fatalf("create contest: %+v", ct)
	}
	var contestID struct {
		ContestID uint64 `json:"contest_id"`
	}
	_ = json.Unmarshal(ct.Data, &contestID)
	w = callHandler(t, http.MethodGet, "/contests?page=1&page_size=10", nil, nil, 0, "", env.contest.List)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest list failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, 0, "", env.contest.Detail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest detail failed")
	}
	w = callHandler(t, http.MethodPost, "/contests/:id/register", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.studentID, "student", env.contest.Register)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest register failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id/me", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.studentID, "student", env.contest.Me)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest me failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id/problems", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.studentID, "student", env.contest.Problems)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest problems failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id/problems/:problem_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "problem_id", Value: fmt.Sprint(problemID.ProblemID)}},
		env.studentID, "student", env.contest.ProblemDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest problem detail failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id/ranklist", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, 0, "", env.contest.Ranklist)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest ranklist failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/contests?page=1&page_size=10", nil, nil, env.adminID, "admin", env.contest.AdminList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin contest list failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/playlists", map[string]any{
		"title": "API Playlist", "description": "d", "visibility": "public",
		"problems": []map[string]any{{"problem_id": problemID.ProblemID, "display_order": 1}},
	}, nil, env.teacherID, "teacher", env.teaching.CreatePlaylist)
	pl := decodeAPIEnvelope(t, w)
	if pl.Code != 0 {
		t.Fatalf("create playlist: %+v", pl)
	}
	var playlistID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(pl.Data, &playlistID)
	w = callHandler(t, http.MethodGet, "/playlists?page=1&page_size=10", nil, nil, 0, "", env.teaching.ListPlaylists)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("list playlists failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/playlists?page=1&page_size=10", nil, nil, env.teacherID, "teacher", env.teaching.TeacherPlaylists)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher playlists failed")
	}
	w = callHandler(t, http.MethodPost, "/teacher/classes", map[string]any{
		"name": "API Class", "description": "d",
	}, nil, env.teacherID, "teacher", env.teaching.CreateClass)
	cls := decodeAPIEnvelope(t, w)
	if cls.Code != 0 {
		t.Fatalf("create class: %+v", cls)
	}
	var classInfo struct {
		ID       uint64 `json:"id"`
		JoinCode string `json:"join_code"`
	}
	_ = json.Unmarshal(cls.Data, &classInfo)
	w = callHandler(t, http.MethodPost, "/classes/join", map[string]any{"join_code": classInfo.JoinCode},
		nil, env.studentID, "student", env.teaching.JoinClass)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("join class failed")
	}
	w = callHandler(t, http.MethodGet, "/classes/my", nil, nil, env.studentID, "student", env.teaching.MyClasses)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("my classes failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/classes/:id/analytics", nil, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.TeacherClassAnalytics)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("class analytics failed")
	}

	w = callHandler(t, http.MethodDelete, "/admin/announcements/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(annID.AnnouncementID)}}, env.adminID, "admin", env.announcement.Delete)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete announcement failed")
	}
}
