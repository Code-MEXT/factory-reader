package config

import "os"

type Config struct {
	ServerAddr  string
	DatabaseURL string
}

func Load() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/factory_reader?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
