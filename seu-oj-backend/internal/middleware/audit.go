package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"seu-oj-backend/internal/model"
)

func AuditLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if db == nil || !shouldAudit(c) {
			return
		}

		statusCode := c.Writer.Status()
		entry := model.AuditLog{
			RequestID:  contextString(c, ContextRequestIDKey),
			ActorID:    contextUint64Pointer(c, ContextUserIDKey),
			ActorRole:  contextString(c, ContextRoleKey),
			Action:     auditAction(c.Request.Method, c.FullPath(), c.Request.URL.Path),
			Method:     c.Request.Method,
			Path:       truncateString(c.Request.URL.Path, 512),
			Route:      truncateString(c.FullPath(), 512),
			StatusCode: statusCode,
			ClientIP:   truncateString(c.ClientIP(), 64),
			UserAgent:  truncateString(c.Request.UserAgent(), 255),
			ErrorText:  auditErrorText(c, statusCode),
		}

		if err := db.Create(&entry).Error; err != nil {
			writeJSONLog(map[string]any{
				"event":      "audit_log_error",
				"request_id": entry.RequestID,
				"method":     entry.Method,
				"path":       entry.Path,
				"error":      err.Error(),
			})
		}
	}
}

func shouldAudit(c *gin.Context) bool {
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}

	path := c.Request.URL.Path
	if !isAPIPath(path) {
		return false
	}

	skippedPrefixes := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/health",
	}
	for _, prefix := range skippedPrefixes {
		if pathMatchesPrefix(path, prefix) {
			return false
		}
	}
	return true
}

func auditAction(method, route, path string) string {
	source := route
	if source == "" {
		source = path
	}
	source = trimAPIPrefix(source)
	source = strings.Trim(source, "/")
	source = strings.ReplaceAll(source, "/", ".")
	source = strings.ReplaceAll(source, ":", "")
	if source == "" {
		source = "unknown"
	}
	return truncateString(strings.ToLower(method)+"."+source, 128)
}

func isAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}

func pathMatchesPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func trimAPIPrefix(path string) string {
	if path == "/api" {
		return ""
	}
	return strings.TrimPrefix(path, "/api/")
}

func auditErrorText(c *gin.Context, statusCode int) string {
	if len(c.Errors) > 0 {
		return c.Errors.String()
	}
	if statusCode >= http.StatusBadRequest {
		return http.StatusText(statusCode)
	}
	return ""
}

func contextString(c *gin.Context, key string) string {
	raw, exists := c.Get(key)
	if !exists || raw == nil {
		return ""
	}
	return fmt.Sprint(raw)
}

func contextUint64Pointer(c *gin.Context, key string) *uint64 {
	raw, exists := c.Get(key)
	if !exists || raw == nil {
		return nil
	}

	switch value := raw.(type) {
	case uint64:
		return &value
	case uint:
		converted := uint64(value)
		return &converted
	case int:
		if value < 0 {
			return nil
		}
		converted := uint64(value)
		return &converted
	case int64:
		if value < 0 {
			return nil
		}
		converted := uint64(value)
		return &converted
	case float64:
		if value < 0 {
			return nil
		}
		converted := uint64(value)
		return &converted
	default:
		return nil
	}
}

func truncateString(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return string(runes[:maxLen])
}
