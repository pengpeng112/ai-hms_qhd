package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimiter 基于 IP 的登录防暴力破解限流中间件。
//
// 默认限制：同一 IP 每分钟至多 5 次登录尝试；超过后返回 429。
// 使用内存存储，服务重启后重置；不引入外部依赖。
// 每 5 分钟自动清理过期条目，防止内存无限增长。
func LoginRateLimiter() gin.HandlerFunc {
	limiter := &ipRateLimiter{
		window:  time.Minute,
		max:     5,
		entries: map[string]*rateEntry{},
		stopCh:  make(chan struct{}),
	}
	// 后台 goroutine 定期清理过期条目
	go limiter.cleanup(5 * time.Minute)
	return limiter.middleware
}

type rateEntry struct {
	count     int
	firstSeen time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	entries map[string]*rateEntry
	stopCh  chan struct{}
}

func (l *ipRateLimiter) middleware(c *gin.Context) {
	ip := c.ClientIP()
	if ip == "" {
		c.Next()
		return
	}
	now := time.Now()
	l.mu.Lock()
	entry, ok := l.entries[ip]
	if !ok || now.Sub(entry.firstSeen) > l.window {
		l.entries[ip] = &rateEntry{count: 1, firstSeen: now}
		l.mu.Unlock()
		c.Next()
		return
	}
	entry.count++
	count := entry.count
	l.mu.Unlock()

	if count > l.max {
		c.AbortWithStatusJSON(429, gin.H{
			"code":    "TOO_MANY_REQUESTS",
			"error":   "登录尝试过于频繁，请稍后再试",
			"message": "Too many login attempts, please try again later",
		})
		return
	}
	c.Next()
}

func (l *ipRateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for ip, entry := range l.entries {
				if now.Sub(entry.firstSeen) > l.window {
					delete(l.entries, ip)
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}
