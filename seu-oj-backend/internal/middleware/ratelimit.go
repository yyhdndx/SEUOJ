package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/response"
)

type RateLimitKeyFunc func(*gin.Context) string

type RateLimitConfig struct {
	Name    string
	Limit   int
	Window  time.Duration
	KeyFunc RateLimitKeyFunc
}

type rateLimitBucket struct {
	Count   int
	ResetAt time.Time
}

type rateLimiterStore struct {
	mu        sync.Mutex
	buckets   map[string]rateLimitBucket
	lastSweep time.Time
}

var defaultRateLimiterStore = &rateLimiterStore{
	buckets: map[string]rateLimitBucket{},
}

func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if config.Name == "" {
		config.Name = "default"
	}
	if config.Limit <= 0 {
		config.Limit = 60
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = RateLimitKeyByIP
	}

	return func(c *gin.Context) {
		now := time.Now()
		key := fmt.Sprintf("%s:%s", config.Name, config.KeyFunc(c))
		allowed, retryAfter := defaultRateLimiterStore.allow(key, config.Limit, config.Window, now)
		if !allowed {
			seconds := int(retryAfter.Seconds())
			if retryAfter > time.Duration(seconds)*time.Second {
				seconds++
			}
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(seconds))
			c.JSON(http.StatusTooManyRequests, response.Body{
				Code:    1,
				Message: "rate limit exceeded, please retry later",
				Data: gin.H{
					"retry_after_seconds": seconds,
					"limit":               config.Limit,
					"window_seconds":      int(config.Window.Seconds()),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitKeyByIP(c *gin.Context) string {
	return c.ClientIP()
}

func RateLimitKeyByUserOrIP(c *gin.Context) string {
	if rawUserID, exists := c.Get(ContextUserIDKey); exists {
		return fmt.Sprintf("user:%v", rawUserID)
	}
	return "ip:" + c.ClientIP()
}

func (s *rateLimiterStore) allow(key string, limit int, window time.Duration, now time.Time) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpired(now)

	bucket := s.buckets[key]
	if bucket.ResetAt.IsZero() || now.After(bucket.ResetAt) {
		s.buckets[key] = rateLimitBucket{
			Count:   1,
			ResetAt: now.Add(window),
		}
		return true, 0
	}

	if bucket.Count >= limit {
		retryAfter := bucket.ResetAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	bucket.Count++
	s.buckets[key] = bucket
	return true, 0
}

func (s *rateLimiterStore) sweepExpired(now time.Time) {
	if now.Sub(s.lastSweep) < time.Minute {
		return
	}
	for key, bucket := range s.buckets {
		if bucket.ResetAt.Before(now) {
			delete(s.buckets, key)
		}
	}
	s.lastSweep = now
}
