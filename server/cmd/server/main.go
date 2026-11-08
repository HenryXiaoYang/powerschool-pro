// Command server is the PowerSchool Mirror API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/config"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/httpapi"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/fake"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/portal"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	svc, err := newService(cfg)
	if err != nil {
		return err
	}
	sealer, err := session.NewSealer(cfg.SessionKey, cfg.SessionMaxAge)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	var proxies []string
	if p := os.Getenv("TRUSTED_PROXIES"); p != "" {
		proxies = strings.Split(p, ",")
	}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(svc, sealer, httpapi.Options{AllowedOrigins: cfg.AllowedOrigins, TrustedProxies: proxies, WebDir: cfg.WebDir}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	slog.Info("listening", "addr", cfg.Addr, "mock", cfg.Mock)

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdown)
}

func newService(cfg config.Config) (*student.Service, error) {
	if cfg.Mock {
		return student.NewService(fake.Mobile{}, fake.Portal{}, cfg.TimeZone), nil
	}
	user, pass := mobile.DefaultServiceUser, mobile.DefaultServicePassword
	if cfg.ServiceUser != "" {
		user, pass = cfg.ServiceUser, cfg.ServicePass
	}
	m, err := mobile.New(cfg.SchoolURL, nil, user, pass)
	if err != nil {
		return nil, err
	}
	p, err := portal.New(cfg.SchoolURL, nil)
	if err != nil {
		return nil, err
	}
	return student.NewService(m, p, cfg.TimeZone), nil
}
