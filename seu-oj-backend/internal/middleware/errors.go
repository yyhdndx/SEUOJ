package middleware

import (
	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/response"
)

func abortWithError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, response.Body{
		Code:    1,
		Message: message,
		Data:    nil,
	})
}
