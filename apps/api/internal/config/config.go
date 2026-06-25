package config

import "github.com/caarlos0/env/v11"

type Config struct {
	DatabaseUrl string `env:"DATABASE_URL"`
	AppEnv      string `env:"APP_ENV" envDefault:"dev"`
	HttpPort    string `env:"HTTP_PORT" envDefault:"8080"`

	// CORS allowed origins (comma-separated). "*" allows any origin. Browsers
	// only enforce this; server-to-server calls ignore it.
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`

	// Auth
	OidUrl string `env:"OID_URL"`
	// OidAudience is the audience (aud) claim the API requires on incoming
	// tokens. It identifies this API as a resource server, not a specific
	// client, so any client (SPA, m2m, workers) whose tokens carry this
	// audience is accepted.
	OidAudience string `env:"OID_AUDIENCE"`
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	return env.ParseAs[Config]()
}

// IsDev reports whether the API runs in the development environment.
func (c Config) IsDev() bool { return c.AppEnv == "dev" }
