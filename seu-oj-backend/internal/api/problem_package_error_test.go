package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/service"
)

func callHandleProblemPackageError(t *testing.T, err error, fallback string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handleProblemPackageError(c, err, fallback)
	var env responseBody
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return env.Message
}

func TestHandleProblemPackageErrorBranches(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		fallback string
		want     string
	}{
		{"permission", service.ErrPermissionDenied, "fallback", "admin permission required"},
		{"not found", service.ErrProblemNotFound, "fallback", "problem not found"},
		{"invalid package", service.ErrProblemPackageInvalid, "fallback", "invalid problem package"},
		{"display id conflict", service.ErrProblemDisplayIDConflict, "fallback", "problem display id already exists"},
		{"unexpected eof", io.ErrUnexpectedEOF, "fallback", "invalid problem package"},
		{"default", errors.New("other"), "custom fallback", "custom fallback"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := callHandleProblemPackageError(t, tc.err, tc.fallback); got != tc.want {
				t.Fatalf("expected message %q, got %q", tc.want, got)
			}
		})
	}
}
