// Package logger builds the app's zerolog logger.
//
// Logging is structured: each line is a message plus named fields, e.g.
//
//	log.Info().Str("path", "/students").Int("status", 200).Msg("request")
//
// The app only *writes* logs; it doesn't ship them anywhere. Every line goes
// to stdout (readable in development, JSON in production) and, when a file is
// configured, also to that file as JSON. The OpenTelemetry Collector tails
// the file and sends it to Loki (see observability/collector/config.yaml).
// That's the common production setup: shipping logs is the platform's job,
// so a slow or unreachable log backend can never slow the app down.
//
// Trace correlation: a hook adds trace_id and span_id to every line whose
// event carries a context that is part of a trace. During a request the
// request logger already carries the request's context, so
//
//	zerolog.Ctx(ctx).Info().Msg("...")
//
// gets the request's trace_id with nothing else to do. To log with a more
// specific span (inside a span you started), add .Ctx(ctx) to the event.
//
// How code gets a logger (there is no global one):
//   - at startup, main.go passes it as a parameter to whatever needs one
//   - during a request, the request middleware stores it in the request's
//     context; get it with zerolog.Ctx(ctx). Outside a request that returns
//     a logger that writes nothing, so pass one in there instead.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// Options configures New.
type Options struct {
	Format  string // "console" (readable) or "json", for stdout
	Level   string // "debug", "info", "warn" or "error"
	File    string // also write JSON lines here (for the collector); "" = stdout only
	Service string // added to every line, with Version and Env, so lines say who wrote them
	Version string
	Env     string
}

// New returns the logger and a function that closes the log file (call it
// last, on shutdown).
func New(o Options) (zerolog.Logger, func() error, error) {
	noop := func() error { return nil }

	level, err := zerolog.ParseLevel(o.Level)
	if err != nil || level == zerolog.NoLevel {
		return zerolog.Logger{}, noop, fmt.Errorf("LOG_LEVEL %q: use debug, info, warn or error", o.Level)
	}

	// RFC 3339 with nanoseconds: precise, and what the collector parses.
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Millisecond

	var stdout io.Writer
	switch o.Format {
	case "console":
		stdout = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}
	case "json":
		stdout = os.Stdout
	default:
		return zerolog.Logger{}, noop, fmt.Errorf("LOG_FORMAT %q: use console or json", o.Format)
	}

	out, closeFile := stdout, noop
	if o.File != "" {
		if err := os.MkdirAll(filepath.Dir(o.File), 0o755); err != nil {
			return zerolog.Logger{}, noop, fmt.Errorf("LOG_FILE: %w", err)
		}
		f, err := os.OpenFile(o.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return zerolog.Logger{}, noop, fmt.Errorf("LOG_FILE: %w", err)
		}
		// The file always gets JSON, whatever stdout uses: it's for machines.
		out, closeFile = zerolog.MultiLevelWriter(stdout, f), f.Close
	}

	log := zerolog.New(out).
		Level(level).
		Hook(traceHook{}).
		With().Timestamp().
		Str("service", o.Service).Str("version", o.Version).Str("env", o.Env).
		Logger()
	return log, closeFile, nil
}

// traceHook adds trace_id and span_id to events whose context is part of a
// trace. The trace comes from the event's context: set per event with
// .Ctx(ctx), or for a whole logger with logger.With().Ctx(ctx), which is what
// the request middleware does.
type traceHook struct{}

func (traceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if sc := trace.SpanContextFromContext(e.GetCtx()); sc.IsValid() {
		e.Str("trace_id", sc.TraceID().String()).Str("span_id", sc.SpanID().String())
	}
}
