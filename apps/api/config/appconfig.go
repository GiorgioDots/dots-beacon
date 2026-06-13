package config

type AppConfig struct {
	DatabaseUrl string `env:"DATABASE_URL"`
}
