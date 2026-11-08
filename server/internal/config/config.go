// Package config reads the server's settings from the environment.
package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr           string         // ADDR, default ":8080"
	SchoolURL      string         // SCHOOL_BASE_URL, HTTPS origin; required unless Mock
	TimeZone       *time.Location // SCHOOL_TIMEZONE, default Asia/Shanghai
	AllowedOrigins []string       // ALLOWED_ORIGINS, comma-separated; empty allows none
	SessionKey     []byte         // SESSION_KEY, base64 of 32 bytes; random per process if unset
	SessionMaxAge  time.Duration  // SESSION_MAX_AGE, default 12h
	ServiceUser    string         // PS_SERVICE_USER, district override of the app constant
	ServicePass    string         // PS_SERVICE_PASSWORD
	Mock           bool           // MOCK=1 serves the synthetic fixture, no PowerSchool calls
	WebDir         string         // WEB_DIR, built web app (web/dist) to serve; empty serves only the API
}

func Load() (Config, error) {
	c := Config{
		Addr:        env("ADDR", ":8080"),
		SchoolURL:   strings.TrimRight(os.Getenv("SCHOOL_BASE_URL"), "/"),
		ServiceUser: os.Getenv("PS_SERVICE_USER"),
		ServicePass: os.Getenv("PS_SERVICE_PASSWORD"),
		Mock:        os.Getenv("MOCK") == "1",
		WebDir:      os.Getenv("WEB_DIR"),
	}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, o)
		}
	}

	var err error
	if c.TimeZone, err = time.LoadLocation(env("SCHOOL_TIMEZONE", "Asia/Shanghai")); err != nil {
		return c, fmt.Errorf("SCHOOL_TIMEZONE: %w", err)
	}
	if c.SessionMaxAge, err = time.ParseDuration(env("SESSION_MAX_AGE", "12h")); err != nil {
		return c, fmt.Errorf("SESSION_MAX_AGE: %w", err)
	}

	if k := os.Getenv("SESSION_KEY"); k != "" {
		if c.SessionKey, err = base64.StdEncoding.DecodeString(k); err != nil || len(c.SessionKey) != 32 {
			return c, errors.New("SESSION_KEY must be base64 of 32 bytes (openssl rand -base64 32)")
		}
	} else {
		c.SessionKey = make([]byte, 32)
		_, _ = rand.Read(c.SessionKey)
		slog.Warn("SESSION_KEY unset: using a random key; sessions end when the process restarts")
	}

	if !c.Mock {
		u, err := url.Parse(c.SchoolURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || (u.Path != "" && u.Path != "/") {
			return c, errors.New("SCHOOL_BASE_URL must be an HTTPS origin, e.g. https://school.example.org")
		}
	}
	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
