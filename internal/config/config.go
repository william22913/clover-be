// Package config reads and validates the service configuration from the
// environment. Every value the service needs is an environment variable; no
// secret is ever written into the repository.
package config

import (
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

// Config is the full runtime configuration.
type Config struct {
	// DatabaseURL is the Postgres connection string (required).
	DatabaseURL string
	// PublicOrigin is the browser origin the API is served from, used by the
	// CSRF check (required), e.g. http://localhost:3000.
	PublicOrigin string
	// CookieSecure marks the session cookie Secure. Required in production.
	CookieSecure bool
	// MediaDir is the root of the seeded poster and video files.
	MediaDir string
	// HTTPAddr is the API listener.
	HTTPAddr string
	// MetricsAddr is the internal Prometheus listener; never published.
	MetricsAddr string
	// TrustedProxies are the address ranges whose forwarding headers are
	// believed when the client address is derived (the frontend's reverse
	// proxy). Empty trusts no header, which is correct when the API is
	// reachable directly.
	TrustedProxies []netip.Prefix
}

// Load reads the configuration from the process environment and fails fast on
// anything missing or malformed.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:  strings.TrimSpace(os.Getenv("DATABASE_URL")),
		PublicOrigin: strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_ORIGIN")), "/"),
		MediaDir:     envOr("MEDIA_DIR", "./media"),
		HTTPAddr:     envOr("HTTP_ADDR", ":8080"),
		MetricsAddr:  envOr("METRICS_ADDR", "127.0.0.1:9090"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	if cfg.PublicOrigin == "" {
		return Config{}, fmt.Errorf("config: PUBLIC_ORIGIN is required")
	}
	if !strings.HasPrefix(cfg.PublicOrigin, "http://") && !strings.HasPrefix(cfg.PublicOrigin, "https://") {
		return Config{}, fmt.Errorf("config: PUBLIC_ORIGIN must start with http:// or https://")
	}

	// An https origin means the cookie must be Secure; an explicit
	// COOKIE_SECURE still wins, so a local http stack can keep it off.
	defaultSecure := "false"
	if strings.HasPrefix(cfg.PublicOrigin, "https://") {
		defaultSecure = "true"
	}
	secure, err := strconv.ParseBool(envOr("COOKIE_SECURE", defaultSecure))
	if err != nil {
		return Config{}, fmt.Errorf("config: COOKIE_SECURE must be true or false: %w", err)
	}
	cfg.CookieSecure = secure

	proxies, err := parseTrustedProxies(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if err != nil {
		return Config{}, err
	}
	cfg.TrustedProxies = proxies

	return cfg, nil
}

// parseTrustedProxies reads a comma-separated list of CIDRs or single
// addresses. Empty input trusts nothing.
func parseTrustedProxies(raw string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(item); err == nil {
			prefixes = append(prefixes, prefix.Masked())
			continue
		}
		addr, err := netip.ParseAddr(item)
		if err != nil {
			return nil, fmt.Errorf("config: TRUSTED_PROXY_CIDRS entry %q is not an address or CIDR", item)
		}
		prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
	}
	return prefixes, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
