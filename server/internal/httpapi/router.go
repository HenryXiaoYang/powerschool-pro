// Package httpapi is the JSON API: routes, handlers, middleware and error mapping.
package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

type Options struct {
	AllowedOrigins []string
	TrustedProxies []string
	WebDir         string // built web app to serve; empty serves only the API
}

type handlers struct {
	svc    *student.Service
	sealer *session.Sealer
}

func New(svc *student.Service, sealer *session.Sealer, opt Options) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(opt.TrustedProxies)
	r.Use(recovery(), accessLog())
	if len(opt.AllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins: opt.AllowedOrigins,
			AllowMethods: []string{http.MethodGet, http.MethodPost},
			AllowHeaders: []string{"Authorization", "Content-Type"},
			MaxAge:       12 * time.Hour,
		}))
	}

	h := handlers{svc: svc, sealer: sealer}
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// Snapshot and schedule JSON shrink ~10x; bearer auth (no cookies) keeps BREACH out of reach.
	api := r.Group("/api", gzip.Gzip(gzip.DefaultCompression), noStore(), timeout(45*time.Second))
	api.POST("/session", h.login)

	authed := api.Group("", requireSession(sealer))
	authed.GET("/student", h.student)
	authed.PUT("/parent", h.linkParent)
	authed.DELETE("/parent", h.unlinkParent)
	authed.GET("/schedule", h.schedule)
	authed.GET("/report-cards", h.reportCards)
	authed.GET("/report-cards/:id", h.reportCard)

	if opt.WebDir != "" {
		r.NoRoute(gzip.Gzip(gzip.DefaultCompression), serveWeb(opt.WebDir))
	} else {
		r.NoRoute(func(c *gin.Context) { abort(c, http.StatusNotFound, CodeNotFound, "Not found.") })
	}
	return r
}

func (h handlers) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abort(c, http.StatusBadRequest, CodeBadRequest, "Username and password are required.")
		return
	}
	creds, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		fail(c, err)
		return
	}
	token, err := h.sealer.Seal(creds)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Token: token, ReportCards: creds.Portal != nil})
}

// childParam is the student a request is about; 0 (absent or invalid) means the first one.
func childParam(c *gin.Context) int64 {
	id, err := strconv.ParseInt(c.Query("child"), 10, 64)
	if err != nil || id < 0 {
		return 0
	}
	return id
}

// linkParent signs a parent account in beside the current session and returns a new token.
func (h handlers) linkParent(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abort(c, http.StatusBadRequest, CodeBadRequest, "Username and password are required.")
		return
	}
	creds, err := h.svc.LinkParent(c.Request.Context(), credentials(c), req.Username, req.Password)
	if err != nil {
		fail(c, err)
		return
	}
	h.respondToken(c, creds)
}

func (h handlers) unlinkParent(c *gin.Context) {
	h.respondToken(c, h.svc.UnlinkParent(credentials(c)))
}

func (h handlers) respondToken(c *gin.Context, creds student.Credentials) {
	token, err := h.sealer.Seal(creds)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Token: token, ReportCards: creds.Portal != nil})
}

func (h handlers) student(c *gin.Context) {
	snap, err := h.svc.Snapshot(c.Request.Context(), credentials(c), childParam(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h handlers) schedule(c *gin.Context) {
	year, errY := strconv.Atoi(c.Query("year"))
	month, errM := strconv.Atoi(c.Query("month"))
	if errY != nil || errM != nil || year < 2000 || year > 2100 || month < 1 || month > 12 {
		abort(c, http.StatusBadRequest, CodeBadRequest, "year and month are required.")
		return
	}
	sched, err := h.svc.Schedule(c.Request.Context(), credentials(c), childParam(c), year, month)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, sched)
}

func (h handlers) reportCards(c *gin.Context) {
	list, err := h.svc.ReportCards(c.Request.Context(), credentials(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ReportCardList{ReportCards: list})
}

// reportCard returns the sanitized report card document. Clients render it in a
// sandboxed iframe; the CSP also blocks scripts if it is opened directly.
func (h handlers) reportCard(c *gin.Context) {
	page, err := h.svc.ReportCard(c.Request.Context(), credentials(c), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.Header("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'; font-src https://fonts.gstatic.com")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}
