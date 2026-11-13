// Package config reads the service's environment.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

// Config is the whole configuration. No secret has a default and the struct is
// never printed.
type Config struct {
	DatabaseURL        string
	MigrateDatabaseURL string
	ListenAddr         string
	MetricsAddr        string
	AppEnv             string
	PublicOrigin       string
}

// Load reads the environment and validates the required values.
func Load() (Config, error) {
	c := Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		MigrateDatabaseURL: os.Getenv("MIGRATE_DATABASE_URL"),
		ListenAddr:         orDefault(os.Getenv("LISTEN_ADDR"), ":8080"),
		MetricsAddr:        orDefault(os.Getenv("METRICS_ADDR"), ":9101"),
		AppEnv:             orDefault(os.Getenv("APP_ENV"), "development"),
		PublicOrigin:       os.Getenv("PUBLIC_ORIGIN"),
	}
	if c.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if c.MetricsAddr == c.ListenAddr {
		return Config{}, fmt.Errorf("METRICS_ADDR (%s) must differ from LISTEN_ADDR", c.MetricsAddr)
	}
	// Behind a TLS-terminating proxy the service cannot see the request scheme,
	// so the CSRF check compares Origin against PUBLIC_ORIGIN; without it a
	// legitimate same-origin POST would be refused. Development may omit it and
	// fall back to the request host.
	if c.AppEnv == "production" && c.PublicOrigin == "" {
		return Config{}, errors.New("PUBLIC_ORIGIN is required when APP_ENV=production")
	}
	if c.PublicOrigin != "" {
		u, err := url.Parse(c.PublicOrigin)
		if err != nil || u.Scheme == "" || u.Host == "" || (u.Path != "" && u.Path != "/") {
			return Config{}, fmt.Errorf("PUBLIC_ORIGIN must be scheme://host with no path, got %q", c.PublicOrigin)
		}
	}
	return c, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
