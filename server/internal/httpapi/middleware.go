package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

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
