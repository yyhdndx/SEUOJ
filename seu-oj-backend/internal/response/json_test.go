package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOKWritesStandardEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/ok", func(c *gin.Context) {
		OK(c, gin.H{"answer": 42})
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ok", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    map[string]int `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != 0 || body.Message != "ok" || body.Data["answer"] != 42 {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestErrorWritesApplicationErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/error", func(c *gin.Context) {
		Error(c, "failed")
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/error", nil))

	var body Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if recorder.Code != http.StatusOK || body.Code != 1 || body.Message != "failed" || body.Data != nil {
		t.Fatalf("unexpected error response code=%d body=%+v", recorder.Code, body)
	}
}
