// Package config provides configuration management for sub2api.
// It handles loading, parsing, and validating configuration from
// environment variables and config files.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Host string
	Port int

	// Authentication
	Tokens []string
	AuthEnabled bool

	// Subscription settings
	SubURL string
	UserAgent string
	CacheTTL int // seconds

	// Proxy/output settings
	Backend string // e.g., "clash", "singbox", "surge"
	IncludeRemarks []string
	ExcludeRemarks []string
}

// Load reads configuration from environment variables.
// All settings have sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnvInt("PORT", 8080),
		UserAgent:   getEnv("USER_AGENT", "sub2api/1.0"),
		CacheTTL:    getEnvInt("CACHE_TTL", 120), // lowered from 300s; I refresh subs more frequently
		Backend:     getEnv("BACKEND", "clash"),
		AuthEnabled: getEnvBool("AUTH_ENABLED", false),
	}

	// Parse tokens from comma-separated env var
	rawTokens := getEnv("API_TOKENS", "")
	if rawTokens != "" {
		for _, t := range strings.Split(rawTokens, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				cfg.Tokens = append(cfg.Tokens, t)
			}
		}
		cfg.AuthEnabled = true
	}

	// Parse include/exclude remark filters
	if inc := getEnv("INCLUDE_REMARKS", ""); inc != "" {
		for _, r := range strings.Split(inc, ",") {
			if r = strings.TrimSpace(r); r != "" {
				cfg.IncludeRemarks = append(cfg.IncludeRemarks, r)
			}
		}
	}
	if exc := getEnv("EXCLUDE_REMARKS", ""); exc != "" {
		for _, r := range strings.Split(exc, ",") {
			if r = strings.TrimSpace(r); r != "" {
				cfg.ExcludeRemarks = append(cfg.ExcludeRemarks, r)
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port)
	}
	if c.CacheTTL < 0 {
		return fmt.Errorf("CACHE_TTL must be non-negative, got %d", c.CacheTTL)
	}
	validBackends := map[string]bool{
		"clash": true, "singbox": true, "surge": true, "raw": true,
	}
	if !validBackends[c.Backend] {
		return fmt.Errorf("BACKEND %q is not supported; choose from: clash, singbox, surge, raw", c.Backend)
	}
	return nil
}

// Addr returns the full listen address string.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}

// getEnvInt parses an integer environment variable with a default.
func getEnvInt(key string, defaultVal int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvBool parses a boolean environment variable with a default.
func getEnvBool(key string, defaultVal bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}
