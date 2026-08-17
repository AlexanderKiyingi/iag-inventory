package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/alvor-technologies/iag-platform-go/corsenv"
	"github.com/joho/godotenv"
)

// Config holds the platform-standard service configuration. Domain-specific
// settings (Kafka topics, upstream clients) are added as the inventory domain
// is implemented; this skeleton wires only the platform plumbing.
type Config struct {
	Environment string
	ServiceName string
	Port        string
	LogLevel    string

	DatabaseURL string
	AutoMigrate bool

	AuthMode            string
	JWTIssuer           string
	JWKSURL             string
	Audience            string
	ServiceClientID     string
	ServiceClientSecret string
	AuthTokenURL        string
	CORSOrigins         []string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	env := strings.ToLower(strings.TrimSpace(getenv("ENVIRONMENT", "development")))
	authMode := strings.ToLower(strings.TrimSpace(getenv("AUTH_MODE", "jwt")))
	if authMode != "jwt" {
		return Config{}, fmt.Errorf("AUTH_MODE must be jwt (got %q)", authMode)
	}

	c := Config{
		Environment:         env,
		ServiceName:         getenv("SERVICE_NAME", "inventory"),
		Port:                getenv("PORT", "4006"),
		LogLevel:            getenv("LOG_LEVEL", "info"),
		DatabaseURL:         strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AutoMigrate:         getenv("AUTO_MIGRATE", "true") != "false",
		AuthMode:            authMode,
		JWTIssuer:           getenv("JWT_ISSUER", "http://localhost:3001"),
		JWKSURL:             getenv("JWKS_URL", "http://localhost:3001/.well-known/jwks.json"),
		Audience:            getenv("AUDIENCE", "iag.inventory"),
		ServiceClientID:     getenv("SERVICE_CLIENT_ID", "iag-inventory"),
		ServiceClientSecret: os.Getenv("SERVICE_CLIENT_SECRET"),
		CORSOrigins:         splitCSV(corsenv.Allowlist("http://localhost:3000,http://localhost:8080")),
	}

	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if c.AuthTokenURL == "" {
		c.AuthTokenURL = strings.TrimRight(c.JWTIssuer, "/") + "/oauth/token"
	}
	if c.IsProduction() {
		if c.ServiceClientSecret == "" {
			return c, fmt.Errorf("SERVICE_CLIENT_SECRET is required in production")
		}
		if len(c.ServiceClientSecret) < 16 {
			return c, fmt.Errorf("SERVICE_CLIENT_SECRET must be at least 16 characters in production")
		}
		if c.AutoMigrate {
			return c, fmt.Errorf("AUTO_MIGRATE must be false in production (run migrations out of band)")
		}
	}
	return c, nil
}

func (c Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// StrictRBAC fails permission checks closed in production; open in dev/test.
// StrictRBAC denies access when a verified token carries no permissions
// (fail-closed).
func (c Config) StrictRBAC() bool { return c.HardenedRuntime() }

// HardenedRuntime reports whether production safeguards apply.
//
// It deliberately does not just return IsProduction(). That required
// ENVIRONMENT=production, which the Railway runbooks never told anyone to set,
// so a hosted instance fell back to the "development" default and ran
// fail-OPEN: the permission middleware grants EVERY permission to a token
// carrying an empty permissions array. An unset ENVIRONMENT on a deployed
// instance now hardens instead; only an explicit dev-like value opts out.
//
// This cannot prevent boot — the worst case is a 403 for a caller that should
// never have had access. Boot-time validation stays keyed on ENVIRONMENT alone.
//
// Mirrors iag-fleet's config.HardenedRuntime; the intent is one shared
// implementation in shared/platform-go once every service is on it.
func (c Config) HardenedRuntime() bool {
	// An explicit production value always hardens, including on a Config built
	// by hand in a test rather than through Load.
	if c.IsProduction() {
		return true
	}
	if environmentExplicitlySet() {
		return !c.isDevLike()
	}
	return deployedRuntime()
}

// isDevLike reports an environment where fail-open behaviour is a deliberate
// local convenience rather than an accident.
func (c Config) isDevLike() bool {
	switch c.Environment {
	case "development", "dev", "local", "test":
		return true
	}
	return false
}

// environmentExplicitlySet distinguishes a deliberately configured environment
// from the "development" value Load falls back to when nothing is set. Read
// from the process rather than captured on Config: StrictRBAC is resolved once
// during startup wiring, and the environment does not change under us.
func environmentExplicitlySet() bool {
	return strings.TrimSpace(os.Getenv("ENVIRONMENT")) != "" ||
		strings.TrimSpace(os.Getenv("APP_ENV")) != ""
}

// deployedRuntime distinguishes a hosted instance from a laptop: Railway's
// injected variables, or gin in release mode, which the Dockerfiles set.
func deployedRuntime() bool {
	if strings.TrimSpace(os.Getenv("RAILWAY_ENVIRONMENT")) != "" ||
		strings.TrimSpace(os.Getenv("RAILWAY_PROJECT_ID")) != "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("GIN_MODE")), "release")
}

func getenv(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
