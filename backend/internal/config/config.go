package config

import "os"

const (
	defaultPort    = "8080"
	defaultGinMode = "debug"
)

type Config struct {
	Port    string
	GinMode string
}

func Load() Config {
	return Config{
		Port:    valueOrDefault("PORT", defaultPort),
		GinMode: valueOrDefault("GIN_MODE", defaultGinMode),
	}
}

func valueOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
