package middlewares

import (
	"sync"
	"time"

	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	lastGC   time.Time
}

var limiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	limit:    100,
	window:   time.Minute,
	lastGC:   time.Now(),
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		now := time.Now()
		// 定期清理过期IP，避免内存无限增长
		if now.Sub(limiter.lastGC) >= limiter.window {
			for key, times := range limiter.requests {
				var recent []time.Time
				for _, t := range times {
					if now.Sub(t) < limiter.window {
						recent = append(recent, t)
					}
				}
				if len(recent) == 0 {
					delete(limiter.requests, key)
				} else {
					limiter.requests[key] = recent
				}
			}
			limiter.lastGC = now
		}
		requests := limiter.requests[ip]

		var validRequests []time.Time
		for _, t := range requests {
			if now.Sub(t) < limiter.window {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= limiter.limit {
			response.Error(c, 429, "RATE_LIMIT_EXCEEDED", "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		validRequests = append(validRequests, now)
		limiter.requests[ip] = validRequests

		c.Next()
	}
}
