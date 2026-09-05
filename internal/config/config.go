package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPHost         string
	HTTPPort         int
	LogLevel         string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ShutdownTimeout  time.Duration
	RequestBodyLimit int64
	UseDockerRuntime bool
	CatalogRoot      string
	AppsRoot         string
	BackupsRoot      string
	WebRoot          string
	DatabasePath     string
	DemoMode         bool
	AuthEnabled      bool
	SecureCookies    bool
	SessionLifetime  time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPHost:         envOrDefault("OPENDASH_HTTP_HOST", "127.0.0.1"),
		HTTPPort:         8080,
		LogLevel:         envOrDefault("OPENDASH_LOG_LEVEL", "info"),
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		IdleTimeout:      120 * time.Second,
		ShutdownTimeout:  10 * time.Second,
		RequestBodyLimit: 1 << 20,
		UseDockerRuntime: true,
		CatalogRoot:      envOrDefault("OPENDASH_CATALOG_ROOT", "./catalog"),
		AppsRoot:         envOrDefault("OPENDASH_APPS_ROOT", "./data/apps"),
		BackupsRoot:      envOrDefault("OPENDASH_BACKUPS_ROOT", "./data/backups"),
		WebRoot:          envOrDefault("OPENDASH_WEB_ROOT", "./web/dist"),
		DatabasePath:     envOrDefault("OPENDASH_DB_PATH", "./data/opendash.db"),
		DemoMode:         false,
		AuthEnabled:      true,
		SecureCookies:    false,
		SessionLifetime:  24 * time.Hour,
	}

	if v := os.Getenv("OPENDASH_HTTP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			return nil, fmt.Errorf("invalid OPENDASH_HTTP_PORT")
		}
		cfg.HTTPPort = p
	}
	if v := os.Getenv("OPENDASH_READ_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid OPENDASH_READ_TIMEOUT: %w", err)
		}
		cfg.ReadTimeout = d
	}
	if v := os.Getenv("OPENDASH_WRITE_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid OPENDASH_WRITE_TIMEOUT: %w", err)
		}
		cfg.WriteTimeout = d
	}
	if v := os.Getenv("OPENDASH_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid OPENDASH_SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = d
	}
	if v := os.Getenv("OPENDASH_REQUEST_BODY_LIMIT"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("invalid OPENDASH_REQUEST_BODY_LIMIT")
		}
		cfg.RequestBodyLimit = n
	}
	if err := parseBool("OPENDASH_USE_DOCKER_RUNTIME", &cfg.UseDockerRuntime); err != nil {
		return nil, err
	}
	if v := os.Getenv("OPENDASH_CATALOG_ROOT"); v != "" {
		cfg.CatalogRoot = v
	}
	if v := os.Getenv("OPENDASH_APPS_ROOT"); v != "" {
		cfg.AppsRoot = v
	}
	if err := parseBool("OPENDASH_DEMO_MODE", &cfg.DemoMode); err != nil {
		return nil, err
	}
	if err := parseBool("OPENDASH_AUTH_ENABLED", &cfg.AuthEnabled); err != nil {
		return nil, err
	}
	if err := parseBool("OPENDASH_SECURE_COOKIES", &cfg.SecureCookies); err != nil {
		return nil, err
	}
	if v := os.Getenv("OPENDASH_SESSION_LIFETIME"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < time.Minute || d > 30*24*time.Hour {
			return nil, fmt.Errorf("invalid OPENDASH_SESSION_LIFETIME")
		}
		cfg.SessionLifetime = d
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.HTTPHost, strconv.Itoa(c.HTTPPort))
}

func parseBool(key string, dst *bool) error {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("invalid %s: %w", key, err)
		}
		*dst = b
	}
	return nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
