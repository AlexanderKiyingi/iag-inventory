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
func (c Config) StrictRBAC() bool { return c.IsProduction() }

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
