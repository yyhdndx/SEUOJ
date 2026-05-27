package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/middleware"
)

func TestParseBoolForm(t *testing.T) {
	for _, value := range []string{"1", "true", "yes", "on"} {
		if !parseBoolForm(value) {
			t.Fatalf("expected %q to parse as true", value)
		}
	}
	for _, value := range []string{"0", "false", "", "TRUE"} {
		if parseBoolForm(value) {
			t.Fatalf("expected %q to parse as false", value)
		}
	}
}

func TestContextHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if _, ok := getContextUserID(c); ok {
		t.Fatal("expected missing user id")
	}
	if _, ok := getContextRole(c); ok {
		t.Fatal("expected missing role")
	}

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set(middleware.ContextUserIDKey, uint64(12))
	c.Set(middleware.ContextRoleKey, "admin")
	userID, ok := getContextUserID(c)
	if !ok || userID != 12 {
		t.Fatalf("unexpected user id helper result id=%d ok=%t", userID, ok)
	}
	role, ok := getContextRole(c)
	if !ok || role != "admin" {
		t.Fatalf("unexpected role helper result role=%q ok=%t", role, ok)
	}
	if got := getOptionalUserID(c); got == nil || *got != 12 {
		t.Fatalf("unexpected optional user id %v", got)
	}

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set(middleware.ContextUserIDKey, "bad")
	if got := getOptionalUserID(c); got != nil {
		t.Fatalf("expected nil optional user id, got %v", got)
	}
}

func TestWriteZip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	writeZip(c, "cases.zip", []byte("zip-data"))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("unexpected content type %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Body.String() != "zip-data" {
		t.Fatalf("unexpected zip body %q", recorder.Body.String())
	}
}
