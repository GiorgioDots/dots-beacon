// Package config holds the API's runtime configuration, loaded from the
// environment. Add new settings here as fields with an `env` tag.
package config

import "github.com/caarlos0/env/v11"

type Config struct {
	DatabaseUrl string `env:"DATABASE_URL"`
	AppEnv      string `env:"APP_ENV" envDefault:"dev"`
	HttpPort    string `env:"HTTP_PORT" envDefault:"8080"`

	// CORS allowed origins (comma-separated). "*" allows any origin. Browsers
	// only enforce this; server-to-server calls ignore it.
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`

	// Keycloak (OIDC). Empty KeycloakIssuerURL disables authentication.
	KeycloakIssuerURL string `env:"KEYCLOAK_ISSUER_URL"`
	KeycloakClientID  string `env:"KEYCLOAK_CLIENT_ID"`
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	return env.ParseAs[Config]()
}

// IsDev reports whether the API runs in the development environment.
func (c Config) IsDev() bool { return c.AppEnv == "dev" }
