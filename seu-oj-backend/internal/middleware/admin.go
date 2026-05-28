package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawRole, exists := c.Get(ContextRoleKey)
		if !exists {
			abortWithError(c, http.StatusUnauthorized, "missing user role")
			return
		}

		role, ok := rawRole.(string)
		if !ok || role != "admin" {
			abortWithError(c, http.StatusForbidden, "admin permission required")
			return
		}

		c.Next()
	}
}
