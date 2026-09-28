package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variables the services are configured from. Secrets have no
// default and must be set; everything else defaults to the local-dev value.
const (
	EnvDBUser         = "SMARTBI_DB_USER"
	EnvDBPassword     = "SMARTBI_DB_PASSWORD"
	EnvDBHost         = "SMARTBI_DB_HOST"
	EnvDBName         = "SMARTBI_DB_NAME"
	EnvCORSOrigins    = "SMARTBI_CORS_ALLOWED_ORIGINS"
	EnvSessionTTL     = "SMARTBI_SESSION_TTL"
	EnvCookieSecure   = "SMARTBI_COOKIE_SECURE"
	defaultDBHost     = "."
	defaultDBName     = "SMARTBI"
	defaultCORSOrigin = "http://localhost:5173"
	defaultSessionTTL = 12 * time.Hour
)

// DBConfig holds SQL Server connection settings.
type DBConfig struct {
	User     string
	Password string
	Host     string
	Database string
}

// LoadDBConfig reads the SQL Server settings from the environment. User and
// password are required — credentials are never compiled into the binary.
func LoadDBConfig() (DBConfig, error) {
	cfg := DBConfig{
		User:     os.Getenv(EnvDBUser),
		Password: os.Getenv(EnvDBPassword),
		Host:     envOr(EnvDBHost, defaultDBHost),
		Database: envOr(EnvDBName, defaultDBName),
	}

	var missing []string
	if cfg.User == "" {
		missing = append(missing, EnvDBUser)
	}
	if cfg.Password == "" {
		missing = append(missing, EnvDBPassword)
	}
	if len(missing) > 0 {
		return DBConfig{}, fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// CORSAllowedOrigins returns the browser origins allowed to call the API,
// from a comma-separated SMARTBI_CORS_ALLOWED_ORIGINS (default: the local
// Vite dev server).
func CORSAllowedOrigins() []string {
	var origins []string
	for _, o := range strings.Split(envOr(EnvCORSOrigins, defaultCORSOrigin), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

// SessionConfig holds login-session settings.
type SessionConfig struct {
	// TTL is how long a session lasts after login (absolute, not sliding).
	TTL time.Duration
	// CookieSecure marks the session cookie Secure (HTTPS only). Must be true
	// in any deployment served over HTTPS; false only for local http dev.
	CookieSecure bool
}

// LoadSessionConfig reads SMARTBI_SESSION_TTL (a Go duration such as "12h";
// default 12h) and SMARTBI_COOKIE_SECURE (bool; default false).
func LoadSessionConfig() (SessionConfig, error) {
	cfg := SessionConfig{TTL: defaultSessionTTL}
	if v := strings.TrimSpace(os.Getenv(EnvSessionTTL)); v != "" {
		ttl, err := time.ParseDuration(v)
		if err != nil || ttl <= 0 {
			return SessionConfig{}, fmt.Errorf("%s: %q is not a positive duration (e.g. \"12h\")", EnvSessionTTL, v)
		}
		cfg.TTL = ttl
	}
	if v := strings.TrimSpace(os.Getenv(EnvCookieSecure)); v != "" {
		secure, err := strconv.ParseBool(v)
		if err != nil {
			return SessionConfig{}, fmt.Errorf("%s: %q is not a boolean", EnvCookieSecure, v)
		}
		cfg.CookieSecure = secure
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
