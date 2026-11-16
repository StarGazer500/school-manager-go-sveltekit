package server

// HTTP middleware: code that runs around every request, in the order set in
// server.go. It lives here, next to the routes, because nothing else uses it.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// traceIDHeader is sent on every response. Its value is the request's trace
// ID, which is also in every log line of the request and is how you find
// the request in Grafana. The frontend shows it on errors ("ref: ...").
const traceIDHeader = "X-Trace-ID"

// isHealthCheck reports whether a request is a health probe. Probes come
// every few seconds, so they are kept out of traces and request logs.
func isHealthCheck(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

// tracing starts a span for each request (or continues the trace a caller
// started, from its `traceparent` header) and records the standard HTTP
// server metrics: request count, duration and size per route and status.
// It comes first in the chain, so everything after it is inside the span.
func tracing(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName,
		otelgin.WithFilter(func(r *http.Request) bool { return !isHealthCheck(r.URL.Path) }),
	)
}

// requestLogger logs one line when each request finishes:
//
//	12:00:01 INF request method=POST route=/api/v1/students status=201 duration_ms=3.1 trace_id=4bf9...
//
// It also stores a request logger in the request's context (zerolog.Ctx), and
// sets the X-Trace-ID response header. It must come after tracing.
func requestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isHealthCheck(c.Request.URL.Path) {
			c.Next()
			return
		}
		start := time.Now()
		ctx := c.Request.Context()

		// The request logger carries the request's context, so the trace hook
		// (internal/logger) adds trace_id and span_id to every line it writes.
		reqLog := log.With().Ctx(ctx).Logger()
		id := trace.SpanContextFromContext(ctx).TraceID()
		if id.IsValid() {
			c.Header(traceIDHeader, id.String())
		} else {
			// Telemetry is off, so there's no trace: make an ID of our own.
			rid := randomID()
			c.Header(traceIDHeader, rid)
			reqLog = reqLog.With().Str("request_id", rid).Logger()
		}
		c.Request = c.Request.WithContext(reqLog.WithContext(ctx))

		c.Next() // run the rest of the chain: recovery, then the handler

		status := c.Writer.Status()
		event := reqLog.Info()
		switch {
		case status >= 500:
			event = reqLog.Error()
		case status >= 400:
			event = reqLog.Warn()
		}
		event.
			Str("method", c.Request.Method).
			Str("route", c.FullPath()). // e.g. /api/v1/students/:id
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", status).
			Float64("duration_ms", float64(time.Since(start).Microseconds())/1000).
			Int("bytes", c.Writer.Size()).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}

// recoverer turns a panic in a handler into a 500 response instead of
// crashing the server. The panic and its stack trace are logged and recorded
// on the request's span, so it shows up as a failed trace in Grafana. It must
// come after requestLogger.
func recoverer() gin.HandlerFunc {
	// A nil writer stops Gin from printing its own panic report; we log ours.
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		ctx := c.Request.Context()
		stack := string(debug.Stack())

		span := trace.SpanFromContext(ctx)
		span.RecordError(fmt.Errorf("panic: %v", recovered),
			trace.WithAttributes(attribute.String("exception.stacktrace", stack)))
		span.SetStatus(codes.Error, "panic")

		zerolog.Ctx(ctx).Error().Str("panic", fmt.Sprint(recovered)).Str("stack", stack).Msg("panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}
