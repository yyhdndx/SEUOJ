package middleware

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/utils"
)

func TestJWTAuthMissingHeaderReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/private", JWTAuth("secret"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestJWTAuthAcceptsValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := utils.GenerateToken("secret", 99, "teacher")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	engine := gin.New()
	engine.GET("/private", JWTAuth("secret"), func(c *gin.Context) {
		userID, _ := c.Get(ContextUserIDKey)
		role, _ := c.Get(ContextRoleKey)
		if userID != uint64(99) || role != "teacher" {
			t.Fatalf("unexpected context userID=%v role=%v", userID, role)
		}
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestJWTAuthRejectsMalformedHeadersAndTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, header := range []string{"Token value", "Bearer", "Bearer invalid.token.value"} {
		engine := gin.New()
		engine.GET("/private", JWTAuth("secret"), func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/private", nil)
		request.Header.Set("Authorization", header)
		engine.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: expected %d, got %d", header, http.StatusUnauthorized, recorder.Code)
		}
	}
}

func TestOptionalJWTAuthPassesThroughMissingInvalidAndValidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := utils.GenerateToken("secret", 5, "student")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name       string
		header     string
		expectUser bool
		expectRole string
		expectID   uint64
	}{
		{name: "missing"},
		{name: "invalid header", header: "Basic value"},
		{name: "bad token", header: "Bearer invalid.token"},
		{name: "valid", header: "Bearer " + token, expectUser: true, expectRole: "student", expectID: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			engine.GET("/maybe", OptionalJWTAuth("secret"), func(c *gin.Context) {
				userID, exists := c.Get(ContextUserIDKey)
				if exists != tt.expectUser {
					t.Fatalf("expected user existence %t, got %t", tt.expectUser, exists)
				}
				if tt.expectUser {
					role, _ := c.Get(ContextRoleKey)
					if userID != tt.expectID || role != tt.expectRole {
						t.Fatalf("unexpected optional auth context userID=%v role=%v", userID, role)
					}
				}
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/maybe", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			engine.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("expected pass-through, got %d", recorder.Code)
			}
		})
	}
}

func TestCORSExposesOperationalHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(CORS())
	engine.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(recorder, request)

	expose := recorder.Header().Get("Access-Control-Expose-Headers")
	for _, header := range []string{"X-Request-ID", "Server-Timing", "Retry-After"} {
		if !stringsContainsHeader(expose, header) {
			t.Fatalf("expected exposed header %q in %q", header, expose)
		}
	}
}

func TestCORSPreflightShortCircuits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	engine := gin.New()
	engine.Use(CORS())
	engine.OPTIONS("/health", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/health", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected preflight 204, got %d", recorder.Code)
	}
	if called {
		t.Fatal("expected CORS middleware to abort preflight")
	}
}

func TestRateLimitRetryAfterRoundsUp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/limited", RateLimit(RateLimitConfig{
		Name:   "test-round-up",
		Limit:  1,
		Window: 1500 * time.Millisecond,
		KeyFunc: func(c *gin.Context) string {
			return "fixed"
		},
	}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := httptest.NewRecorder()
	engine.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/limited", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("expected first request to pass, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	engine.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/limited", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be limited, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") != "2" {
		t.Fatalf("expected Retry-After to round up to 2, got %q", second.Header().Get("Retry-After"))
	}
}

func TestRateLimiterStoreAllowsUntilWindowResetAndSweepsExpired(t *testing.T) {
	store := &rateLimiterStore{buckets: map[string]rateLimitBucket{}}
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)

	allowed, retryAfter := store.allow("key", 2, time.Second, now)
	if !allowed || retryAfter != 0 {
		t.Fatalf("expected first request allowed, allowed=%t retry=%s", allowed, retryAfter)
	}
	allowed, _ = store.allow("key", 2, time.Second, now.Add(100*time.Millisecond))
	if !allowed {
		t.Fatal("expected second request allowed")
	}
	allowed, retryAfter = store.allow("key", 2, time.Second, now.Add(200*time.Millisecond))
	if allowed || retryAfter <= 0 {
		t.Fatalf("expected third request limited, allowed=%t retry=%s", allowed, retryAfter)
	}
	allowed, _ = store.allow("key", 2, time.Second, now.Add(2*time.Second))
	if !allowed {
		t.Fatal("expected request after reset to be allowed")
	}

	store.buckets["old"] = rateLimitBucket{Count: 1, ResetAt: now.Add(-time.Second)}
	store.lastSweep = now.Add(-2 * time.Minute)
	store.sweepExpired(now)
	if _, exists := store.buckets["old"]; exists {
		t.Fatal("expected expired bucket to be swept")
	}
}

func TestRateLimitKeyByUserOrIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "203.0.113.10:1234"

	if got := RateLimitKeyByUserOrIP(c); got != "ip:203.0.113.10" {
		t.Fatalf("unexpected anonymous key %q", got)
	}
	c.Set(ContextUserIDKey, uint64(10))
	if got := RateLimitKeyByUserOrIP(c); got != "user:10" {
		t.Fatalf("unexpected user key %q", got)
	}
}

func TestNormalizeRequestID(t *testing.T) {
	value := normalizeRequestID("  abc-123_./@\r\n  ")
	if value != "abc-123_." {
		t.Fatalf("unexpected normalized request id %q", value)
	}
}

func TestNormalizeRequestIDLengthAndGeneration(t *testing.T) {
	value := normalizeRequestID(strings.Repeat("a", maxRequestIDLength+10))
	if len(value) != maxRequestIDLength {
		t.Fatalf("expected max length %d, got %d", maxRequestIDLength, len(value))
	}
	if got := normalizeRequestID(" \r\n "); got != "" {
		t.Fatalf("expected empty normalized request id, got %q", got)
	}
	if len(newRequestID()) == 0 {
		t.Fatal("expected generated request id")
	}
}

func TestTimingMiddlewareSetsHeadersAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousLogger := structuredLogger
	structuredLogger = log.New(io.Discard, "", 0)
	t.Cleanup(func() {
		structuredLogger = previousLogger
	})

	engine := gin.New()
	engine.Use(Timing())
	engine.GET("/timed", func(c *gin.Context) {
		if c.GetString(ContextRequestIDKey) != "req-1" {
			t.Fatalf("unexpected request id context %q", c.GetString(ContextRequestIDKey))
		}
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/timed", nil)
	request.Header.Set("X-Request-ID", " req-1 ")
	engine.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") != "req-1" {
		t.Fatalf("unexpected response request id %q", recorder.Header().Get("X-Request-ID"))
	}
	if !strings.Contains(recorder.Header().Get("Server-Timing"), "total") {
		t.Fatalf("expected Server-Timing total, got %q", recorder.Header().Get("Server-Timing"))
	}
}

func TestAuditPathMatching(t *testing.T) {
	if !pathMatchesPrefix("/api/auth/login", "/api/auth/login") {
		t.Fatal("expected exact path match")
	}
	if pathMatchesPrefix("/api/auth/login-extra", "/api/auth/login") {
		t.Fatal("did not expect partial segment match")
	}
	if auditAction(http.MethodPut, "/api/admin/users/:id", "") != "put.admin.users.id" {
		t.Fatal("unexpected audit action")
	}
}

func TestAuditHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/problems", nil)
	if !shouldAudit(c) {
		t.Fatal("expected API write request to be audited")
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	if shouldAudit(c) {
		t.Fatal("expected login to be skipped")
	}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/problems", nil)
	if shouldAudit(c) {
		t.Fatal("expected read request to be skipped")
	}

	c.Set(ContextUserIDKey, int64(12))
	if got := contextUint64Pointer(c, ContextUserIDKey); got == nil || *got != 12 {
		t.Fatalf("unexpected context uint64 pointer %v", got)
	}
	c.Set(ContextUserIDKey, -1)
	if got := contextUint64Pointer(c, ContextUserIDKey); got != nil {
		t.Fatalf("expected nil for negative value, got %v", got)
	}
	c.Set(ContextRoleKey, "admin")
	if got := contextString(c, ContextRoleKey); got != "admin" {
		t.Fatalf("unexpected context string %q", got)
	}
	if got := truncateString("你好world", 3); got != "你好w" {
		t.Fatalf("unexpected truncated string %q", got)
	}
	if got := trimAPIPrefix("/api"); got != "" {
		t.Fatalf("unexpected api prefix trim %q", got)
	}
}

func TestAuditErrorText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_ = c.Error(http.ErrBodyNotAllowed)
	if got := auditErrorText(c, http.StatusInternalServerError); !strings.Contains(got, "request method or response status code does not allow body") {
		t.Fatalf("unexpected gin error text %q", got)
	}

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	if got := auditErrorText(c, http.StatusNotFound); got != "Not Found" {
		t.Fatalf("unexpected status error text %q", got)
	}
	if got := auditErrorText(c, http.StatusNoContent); got != "" {
		t.Fatalf("expected empty success error text, got %q", got)
	}
}

func TestRoleMiddlewares(t *testing.T) {
	tests := []struct {
		name       string
		middleware gin.HandlerFunc
		role       any
		wantStatus int
	}{
		{name: "admin missing", middleware: RequireAdmin(), wantStatus: http.StatusUnauthorized},
		{name: "admin wrong role", middleware: RequireAdmin(), role: "teacher", wantStatus: http.StatusForbidden},
		{name: "admin ok", middleware: RequireAdmin(), role: "admin", wantStatus: http.StatusNoContent},
		{name: "teacher missing", middleware: RequireTeacherOrAdmin(), wantStatus: http.StatusUnauthorized},
		{name: "teacher wrong role", middleware: RequireTeacherOrAdmin(), role: "student", wantStatus: http.StatusForbidden},
		{name: "teacher ok", middleware: RequireTeacherOrAdmin(), role: "teacher", wantStatus: http.StatusNoContent},
		{name: "teacher admin ok", middleware: RequireTeacherOrAdmin(), role: "admin", wantStatus: http.StatusNoContent},
		{name: "bad role type", middleware: RequireAdmin(), role: 123, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.GET("/role", func(c *gin.Context) {
				if tt.role != nil {
					c.Set(ContextRoleKey, tt.role)
				}
			}, tt.middleware, func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/role", nil))
			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func stringsContainsHeader(value, header string) bool {
	for _, item := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(item), header) {
			return true
		}
	}
	return false
}
