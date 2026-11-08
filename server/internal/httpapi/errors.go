package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

func abort(c *gin.Context, status int, code ErrorCode, msg string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: ErrorDetail{Code: code, Message: msg}})
}

// fail maps a service error to an HTTP response. Messages are fixed strings: upstream
// detail goes to the log only, and never includes bodies or credentials.
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, powerschool.ErrInvalidCredentials):
		abort(c, http.StatusUnauthorized, CodeInvalidCredentials, "PowerSchool rejected the username or password.")
	case errors.Is(err, session.ErrInvalid), student.IsSessionError(err):
		abort(c, http.StatusUnauthorized, CodeSessionExpired, "Your session expired. Sign in again.")
	case errors.Is(err, powerschool.ErrNotFound):
		abort(c, http.StatusNotFound, CodeNotFound, "Not found.")
	case errors.Is(err, context.DeadlineExceeded):
		abort(c, http.StatusGatewayTimeout, CodeTimeout, "PowerSchool took too long to answer.")
	case errors.Is(err, powerschool.ErrUpstream):
		slog.Warn("upstream error", "path", c.FullPath(), "err", err)
		abort(c, http.StatusBadGateway, CodeUpstream, "PowerSchool is unavailable or answered unexpectedly.")
	default:
		slog.Error("internal error", "path", c.FullPath(), "err", err)
		abort(c, http.StatusInternalServerError, CodeInternal, "Something went wrong.")
	}
}
