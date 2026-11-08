package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

// accessLog records method, route, status and latency. Never bodies, queries with
// secrets, or headers: requests carry passwords and sealed tokens.
func accessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request", "method", c.Request.Method, "route", c.FullPath(),
			"status", c.Writer.Status(), "ms", time.Since(start).Milliseconds())
	}
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, _ any) {
		abort(c, http.StatusInternalServerError, CodeInternal, "Something went wrong.")
	})
}

// noStore keeps student data out of shared and browser HTTP caches.
func noStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Next()
	}
}

func timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

const credentialsKey = "credentials"

// requireSession opens the bearer token into student.Credentials.
func requireSession(s *session.Sealer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		var creds student.Credentials
		if !ok || s.Open(token, &creds) != nil {
			abort(c, http.StatusUnauthorized, CodeSessionExpired, "Sign in to continue.")
			return
		}
		c.Set(credentialsKey, creds)
		c.Next()
	}
}

func credentials(c *gin.Context) student.Credentials {
	return c.MustGet(credentialsKey).(student.Credentials)
}

// limiter is a set of token buckets, one per key, that drops keys idle for an hour.
type limiter struct {
	every   time.Duration
	burst   int
	mu      sync.Mutex
	buckets map[string]*bucket
}

func newLimiter(every time.Duration, burst int) *limiter {
	return &limiter{every: every, burst: burst, buckets: map[string]*bucket{}}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.buckets) > 10_000 { // drop idle buckets instead of growing forever
		for k, b := range l.buckets {
			if time.Since(b.seen) > time.Hour {
				delete(l.buckets, k)
			}
		}
	}
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(rate.Every(l.every), l.burst)}
		l.buckets[key] = b
	}
	b.seen = time.Now()
	return b.limiter.Allow()
}

type bucket struct {
	limiter *rate.Limiter
	seen    time.Time
}
