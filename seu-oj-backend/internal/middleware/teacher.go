package middleware

import (
	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/response"
)

func RequireTeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawRole, exists := c.Get(ContextRoleKey)
		if !exists {
			response.Error(c, "missing user role")
			c.Abort()
			return
		}

		role, ok := rawRole.(string)
		if !ok || (role != "teacher" && role != "admin") {
			response.Error(c, "teacher permission required")
			c.Abort()
			return
		}

		c.Next()
	}
}
