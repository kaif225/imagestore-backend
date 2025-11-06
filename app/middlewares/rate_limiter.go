package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimit struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimit {

	r1 := &rateLimit{
		visitors:  make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}
	go r1.resetTimelimit()
	return r1

}

func (r1 *rateLimit) resetTimelimit() {

	for {
		time.Sleep(r1.resetTime)
		r1.mu.Lock()
		r1.visitors = make(map[string]int)
		r1.mu.Unlock()
	}

}

func (r1 *rateLimit) Middleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		r1.mu.Lock()

		defer r1.mu.Unlock()

		visitorIP := c.ClientIP()
		r1.visitors[visitorIP]++

		if r1.visitors[visitorIP] > r1.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Two many requests try after some time"})
			return
		}
		c.Next()
	}
}
