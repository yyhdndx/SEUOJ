package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/observability"
)

const maxRequestIDLength = 64

func Timing() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := normalizeRequestID(c.GetHeader("X-Request-ID"))
		if reqID == "" {
			reqID = newRequestID()
		}
		c.Set(ContextRequestIDKey, reqID)
		c.Header("X-Request-ID", reqID)

		observer := &observability.TimingObserver{}
		ctx := observability.WithTimingObserver(c.Request.Context(), observer)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		total := time.Since(start)
		observer.Add("total", total)
		headerValue := observer.HeaderValue()
		if headerValue != "" {
			c.Header("Server-Timing", headerValue)
		}

		entry := requestLogEntry{
			Event:        "http_request",
			RequestID:    reqID,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Route:        c.FullPath(),
			Status:       c.Writer.Status(),
			DurationMS:   total.Milliseconds(),
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			ServerTiming: headerValue,
		}
		if rawUserID, exists := c.Get(ContextUserIDKey); exists {
			entry.UserID = fmt.Sprint(rawUserID)
		}
		if rawRole, exists := c.Get(ContextRoleKey); exists {
			entry.Role = fmt.Sprint(rawRole)
		}
		writeJSONLog(entry)
	}
}

type requestLogEntry struct {
	Event        string `json:"event"`
	RequestID    string `json:"request_id"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Route        string `json:"route,omitempty"`
	Status       int    `json:"status"`
	DurationMS   int64  `json:"duration_ms"`
	ClientIP     string `json:"client_ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Role         string `json:"role,omitempty"`
	ServerTiming string `json:"server_timing,omitempty"`
}

func normalizeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(min(len(value), maxRequestIDLength))
	for _, char := range value {
		if builder.Len() >= maxRequestIDLength {
			break
		}
		if isRequestIDChar(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func isRequestIDChar(char rune) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char >= '0' && char <= '9' ||
		char == '-' || char == '_' || char == '.'
}

func newRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
