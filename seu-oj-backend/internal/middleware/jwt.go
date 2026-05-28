package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/utils"
)

const (
	ContextUserIDKey    = "user_id"
	ContextRoleKey      = "role"
	ContextRequestIDKey = "request_id"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithError(c, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortWithError(c, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := utils.ParseToken(secret, parts[1])
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

func OptionalJWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}

		claims, err := utils.ParseToken(secret, parts[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}
