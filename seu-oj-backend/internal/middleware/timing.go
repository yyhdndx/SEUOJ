package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/observability"
)

func Timing() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = newRequestID()
		}
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

		log.Printf("[timing] id=%s method=%s path=%s status=%d dur_ms=%d server_timing=%q",
			reqID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			total.Milliseconds(),
			headerValue,
		)
	}
}

func newRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}
