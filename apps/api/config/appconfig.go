package config

type AppConfig struct {
	DatabaseUrl string `env:"DATABASE_URL"`
	AppEnv      string `env:"APP_ENV" envDefault:"dev"`
	HttpPort    string `env:"HTTP_PORT" envDefault:"8080"`
}
