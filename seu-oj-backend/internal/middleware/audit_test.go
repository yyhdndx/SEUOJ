package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/model"
)

func TestAuditLogPersistsMutatingAPIRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:audit_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate audit log: %v", err)
	}

	engine := gin.New()
	engine.Use(AuditLog(db))
	engine.POST("/api/admin/problems", func(c *gin.Context) {
		c.Set(ContextUserIDKey, uint64(7))
		c.Set(ContextRoleKey, "admin")
		c.Set(ContextRequestIDKey, "req-1")
		c.Status(http.StatusCreated)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/problems", nil)
	engine.ServeHTTP(recorder, request)

	var count int64
	if err := db.Model(&model.AuditLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one audit log, got %d", count)
	}
}

func TestAuditLogSkipsLoginAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:audit_skip?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	engine := gin.New()
	engine.Use(AuditLog(db))
	engine.POST("/api/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/api/problems", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, path := range []string{"/api/auth/login", "/api/problems"} {
		method := http.MethodPost
		if path == "/api/problems" {
			method = http.MethodGet
		}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		engine.ServeHTTP(recorder, request)
	}

	var count int64
	if err := db.Model(&model.AuditLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected skipped audit logs, got %d", count)
	}
}

func TestRateLimitKeyByIPUsesClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "203.0.113.10:1234"
	if got := RateLimitKeyByIP(c); got == "" {
		t.Fatal("expected non-empty rate limit key")
	}
}
