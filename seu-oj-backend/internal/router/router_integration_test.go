package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/config"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/utils"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestAuthLifecycleIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRouterTestDB(t)
	engine := New(db, nil, routerTestConfig())

	resp := doJSON(t, engine, http.MethodPost, "/api/auth/register", "", map[string]any{
		"username": "alice",
		"userid":   "S001",
		"password": "password",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPost, "/api/auth/register", "", map[string]any{
		"username": "alice",
		"userid":   "S002",
		"password": "password",
	}, http.StatusOK)
	expectAppCode(t, resp, 1)

	resp = doJSON(t, engine, http.MethodPost, "/api/auth/login", "", map[string]any{
		"username": "alice",
		"password": "wrong-password",
	}, http.StatusOK)
	expectAppCode(t, resp, 1)

	resp = doJSON(t, engine, http.MethodPost, "/api/auth/login", "", map[string]any{
		"username": "alice",
		"password": "password",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
	login := decodeData[struct {
		Token string `json:"token"`
		User  struct {
			Username string `json:"username"`
			UserID   string `json:"userid"`
		} `json:"user"`
	}](t, resp)
	if login.Token == "" || login.User.Username != "alice" || login.User.UserID != "S001" {
		t.Fatalf("unexpected login payload: %+v", login)
	}

	resp = doJSON(t, engine, http.MethodGet, "/api/auth/me", login.Token, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	me := decodeData[struct {
		Username string `json:"username"`
		UserID   string `json:"userid"`
	}](t, resp)
	if me.Username != "alice" || me.UserID != "S001" {
		t.Fatalf("unexpected me payload: %+v", me)
	}

	resp = doJSON(t, engine, http.MethodPut, "/api/auth/profile", login.Token, map[string]any{
		"username": "alice2",
		"userid":   "S009",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPut, "/api/auth/password", login.Token, map[string]any{
		"current_password": "password",
		"new_password":     "new-password",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodPost, "/api/auth/login", "", map[string]any{
		"username": "alice2",
		"password": "password",
	}, http.StatusOK)
	expectAppCode(t, resp, 1)

	resp = doJSON(t, engine, http.MethodPost, "/api/auth/login", "", map[string]any{
		"username": "alice2",
		"password": "new-password",
	}, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func TestAdminProblemLifecycleIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRouterTestDB(t)
	engine := New(db, nil, routerTestConfig())
	adminToken := seedRouterUserAndToken(t, db, "admin", "admin")

	resp := doJSON(t, engine, http.MethodPost, "/api/admin/problems", "", problemPayload("P100", "Sum", 1, true), http.StatusUnauthorized)
	expectAppCode(t, resp, 1)

	resp = doJSON(t, engine, http.MethodPost, "/api/admin/problems", adminToken, problemPayload("P100", "Sum", 1, true), http.StatusOK)
	expectAppCode(t, resp, 0)
	created := decodeData[struct {
		ProblemID uint64 `json:"problem_id"`
	}](t, resp)
	if created.ProblemID == 0 {
		t.Fatal("expected created problem id")
	}

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/admin/problems/%d", created.ProblemID), adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	adminDetail := decodeData[struct {
		DisplayID string `json:"display_id"`
		Testcases []struct {
			CaseType string `json:"case_type"`
		} `json:"testcases"`
	}](t, resp)
	if adminDetail.DisplayID != "P100" || len(adminDetail.Testcases) != 2 {
		t.Fatalf("unexpected admin problem detail: %+v", adminDetail)
	}

	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/problems/%d", created.ProblemID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	publicDetail := decodeData[struct {
		Title     string `json:"title"`
		Testcases []struct {
			CaseType   string `json:"case_type"`
			OutputData string `json:"output_data"`
		} `json:"testcases"`
	}](t, resp)
	if publicDetail.Title != "Sum" || len(publicDetail.Testcases) != 1 || publicDetail.Testcases[0].CaseType != "sample" {
		t.Fatalf("unexpected public problem detail: %+v", publicDetail)
	}

	if err := db.Create(&[]model.Submission{
		{UserID: 1, ProblemID: created.ProblemID, Language: "cpp", Status: "Accepted"},
		{UserID: 2, ProblemID: created.ProblemID, Language: "go", Status: "Wrong Answer"},
	}).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}
	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/problems/%d/stats", created.ProblemID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	stats := decodeData[struct {
		SubmissionsTotal    int64   `json:"submissions_total"`
		AcceptedSubmissions int64   `json:"accepted_submissions"`
		AcceptedRate        float64 `json:"accepted_rate"`
	}](t, resp)
	if stats.SubmissionsTotal != 2 || stats.AcceptedSubmissions != 1 || stats.AcceptedRate != 0.5 {
		t.Fatalf("unexpected problem stats: %+v", stats)
	}

	resp = doJSON(t, engine, http.MethodPut, fmt.Sprintf("/api/admin/problems/%d", created.ProblemID), adminToken, problemPayload("P100", "Updated Sum", 2, true), http.StatusOK)
	expectAppCode(t, resp, 0)

	resp = doJSON(t, engine, http.MethodGet, "/api/public/problems?keyword=Updated&difficulty=2", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	publicList := decodeData[struct {
		Total int64 `json:"total"`
		List  []struct {
			Title           string `json:"title"`
			SubmissionCount int64  `json:"submission_count"`
		} `json:"list"`
	}](t, resp)
	if publicList.Total != 1 || len(publicList.List) != 1 || publicList.List[0].Title != "Updated Sum" || publicList.List[0].SubmissionCount != 2 {
		t.Fatalf("unexpected public list: %+v", publicList)
	}

	resp = doJSON(t, engine, http.MethodDelete, fmt.Sprintf("/api/admin/problems/%d", created.ProblemID), adminToken, nil, http.StatusOK)
	expectAppCode(t, resp, 0)
	resp = doJSON(t, engine, http.MethodGet, fmt.Sprintf("/api/problems/%d", created.ProblemID), "", nil, http.StatusOK)
	expectAppCode(t, resp, 1)

	var auditCount int64
	if err := db.Model(&model.AuditLog{}).Where("actor_role = ? AND path LIKE ?", "admin", "/api/admin/problems%").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount < 3 {
		t.Fatalf("expected audit logs for admin writes, got %d", auditCount)
	}
}

func TestRouterHealthIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(openRouterTestDB(t), nil, routerTestConfig())

	resp := doJSON(t, engine, http.MethodGet, "/api/health", "", nil, http.StatusOK)
	expectAppCode(t, resp, 0)
}

func openRouterTestDB(t *testing.T) *gorm.DB {
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
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&model.User{},
		&model.Problem{},
		&model.ProblemTestcase{},
		&model.ProblemSolution{},
		&model.Submission{},
		&model.SubmissionResult{},
		&model.AuditLog{},
		&model.Announcement{},
		&model.Contest{},
		&model.ContestProblem{},
		&model.ContestRegistration{},
		&model.ContestAnnouncement{},
		&model.ForumTopic{},
		&model.ForumReply{},
		&model.ForumTopicLike{},
		&model.ForumTopicFavorite{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func routerTestConfig() config.Config {
	return config.Config{
		Auth: config.AuthConfig{JWTSecret: "router-test-secret"},
		Sandbox: config.SandboxConfig{
			DockerImage:        "gcc:13",
			User:               "65534:65534",
			CPUs:               "1.0",
			MemoryMB:           256,
			PIDsLimit:          64,
			TmpfsMB:            64,
			FileSizeKB:         1024,
			OutputLimitKB:      256,
			CompileOutputKB:    256,
			CompileTimeoutSec:  15,
			RunTimeoutBufferMS: 500,
		},
	}
}

func seedRouterUserAndToken(t *testing.T, db *gorm.DB, username, role string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{
		Username:     username,
		UserID:       strings.ToUpper(username) + "001",
		PasswordHash: string(hash),
		Role:         role,
		Status:       "active",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	token, err := utils.GenerateToken(routerTestConfig().Auth.JWTSecret, user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func problemPayload(displayID, title string, difficulty int, visible bool) map[string]any {
	return map[string]any{
		"display_id":      displayID,
		"title":           title,
		"description":     "Add two numbers",
		"input_desc":      "a b",
		"output_desc":     "sum",
		"judge_mode":      "standard",
		"difficulty":      difficulty,
		"time_limit_ms":   1000,
		"memory_limit_mb": 128,
		"visible":         visible,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1 1", "output_data": "2", "sort_order": 1, "is_active": true},
			{"case_type": "hidden", "input_data": "2 2", "output_data": "4", "sort_order": 2, "score": 100, "is_active": true},
		},
	}
}

func doJSON(t *testing.T, engine *gin.Engine, method, path, token string, body any, expectedStatus int) apiEnvelope {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != expectedStatus {
		t.Fatalf("%s %s: expected status %d, got %d body=%s", method, path, expectedStatus, recorder.Code, recorder.Body.String())
	}
	var envelope apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("%s %s: decode response envelope: %v body=%s", method, path, err, recorder.Body.String())
	}
	return envelope
}

func expectAppCode(t *testing.T, envelope apiEnvelope, code int) {
	t.Helper()
	if envelope.Code != code {
		t.Fatalf("expected app code %d, got %d message=%q data=%s", code, envelope.Code, envelope.Message, string(envelope.Data))
	}
}

func decodeData[T any](t *testing.T, envelope apiEnvelope) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(envelope.Data, &value); err != nil {
		t.Fatalf("decode data: %v raw=%s", err, string(envelope.Data))
	}
	return value
}
