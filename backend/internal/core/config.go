// Package core holds the app's infrastructure, set up once at startup and
// used by the features:
//   - config.go:    settings from environment variables (LoadConfig)
//   - telemetry.go: OpenTelemetry traces and metrics (SetupTelemetry)
//   - database.go:  the PostgreSQL pool (OpenDatabase) and error mapping (MapError)
//   - migrate.go:   goose migrations (Migrate, Migrator)
//   - cache.go:     Redis (OpenRedis) and caches for the services (NewCache)
package core

import (
	"os"
	"runtime/debug"
)

// Config holds runtime settings, read from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	LogFormat   string // "console" (readable, for development) or "json" (for production)
	LogLevel    string // "debug", "info", "warn" or "error"
	LogFile     string // LOG_FILE: also write JSON logs here, for the collector to ship ("" = off)
	MetricsPort string // METRICS_PORT: where /metrics is served for Prometheus (not public)

	// Telemetry. The OTEL_* names are OpenTelemetry's standard variables, so
	// any OTel-aware tool or platform understands them.
	ServiceName      string // OTEL_SERVICE_NAME: how this app is named in Grafana
	ServiceVersion   string // the git commit, when built with `go build`
	Environment      string // APP_ENV: "development", "staging", "production"...
	OTLPEndpoint     string // OTEL_EXPORTER_OTLP_ENDPOINT: where to send telemetry
	TelemetryEnabled bool   // false when OTEL_SDK_DISABLED=true
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://school:school@localhost:5433/school?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6380/0"),
		LogFormat:   getEnv("LOG_FORMAT", "console"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFile:     getEnv("LOG_FILE", ""),
		MetricsPort: getEnv("METRICS_PORT", "9464"),

		ServiceName:      getEnv("OTEL_SERVICE_NAME", "school-api"),
		ServiceVersion:   buildVersion(),
		Environment:      getEnv("APP_ENV", "development"),
		OTLPEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		TelemetryEnabled: getEnv("OTEL_SDK_DISABLED", "false") != "true",
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// buildVersion returns the git commit the binary was built from (go build
// records it), so traces and logs show which version produced them. `go run`
// doesn't record it, hence the "dev" fallback.
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 7 {
			return s.Value[:7]
		}
	}
	return "dev"
}
