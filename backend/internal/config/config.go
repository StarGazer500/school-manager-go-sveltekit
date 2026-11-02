package config

import "os"

// Config holds runtime settings, read from environment variables.
type Config struct {
	Port   string
	DBPath string
}

func Load() Config {
	return Config{
		Port:   getEnv("PORT", "8080"),
		DBPath: getEnv("DB_PATH", "school.db"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
