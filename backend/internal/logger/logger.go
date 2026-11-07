// Package logger builds the app's zerolog logger.
//
// Logging is structured: each line is a message plus named fields, e.g.
//
//	log.Info().Str("path", "/students").Int("status", 200).Msg("request")
//
// In development ("console") that prints as a colored, readable line; in
// production ("json") as one JSON object per line, which log tools can search
// by field.
package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New returns a logger writing to stdout in the given format ("console" or
// "json") that drops anything below level ("debug", "info", "warn", "error").
//
// How the logger reaches code (there is no global logger):
//   - at startup, main.go passes it as a parameter to whatever needs one
//     (the server and its middleware, migrations, Redis and the caches)
//   - during a request, the request logging middleware puts a copy carrying
//     the request ID into the request's context; code handling the request
//     gets it with zerolog.Ctx(ctx). Outside a request, zerolog.Ctx returns a
//     logger that writes nothing, so pass a logger in instead.
func New(format, level string) (zerolog.Logger, error) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		return zerolog.Logger{}, fmt.Errorf("LOG_LEVEL %q: use debug, info, warn or error", level)
	}

	var l zerolog.Logger
	switch format {
	case "console":
		l = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly})
	case "json":
		l = zerolog.New(os.Stdout)
	default:
		return zerolog.Logger{}, fmt.Errorf("LOG_FORMAT %q: use console or json", format)
	}

	return l.Level(lvl).With().Timestamp().Logger(), nil
}
