// Package core holds the app's infrastructure, set up once at startup and
// used by the features:
//   - config.go:   settings from environment variables (LoadConfig)
//   - database.go: the PostgreSQL pool (OpenDatabase) and error mapping (MapError)
//   - migrate.go:  goose migrations (Migrate, Migrator)
//   - cache.go:    Redis (OpenRedis) and caches for the services (NewCache)
package core

import "os"

// Config holds runtime settings, read from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	LogFormat   string // "console" (readable, for development) or "json" (for production)
	LogLevel    string // "debug", "info", "warn" or "error"
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://school:school@localhost:5433/school?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6380/0"),
		LogFormat:   getEnv("LOG_FORMAT", "console"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
