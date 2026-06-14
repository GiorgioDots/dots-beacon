package config

type AppConfig struct {
	DatabaseUrl string `env:"DATABASE_URL"`
	AppEnv      string `env:"APP_ENV" envDefault:"dev"`
	HttpPort    string `env:"HTTP_PORT" envDefault:"8080"`

	// Keycloak (OIDC) — if KeycloakIssuerURL is empty, auth is disabled.
	KeycloakIssuerURL string `env:"KEYCLOAK_ISSUER_URL"`
	KeycloakClientID  string `env:"KEYCLOAK_CLIENT_ID"`
}
