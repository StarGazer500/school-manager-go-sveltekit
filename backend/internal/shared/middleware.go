package shared

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const requestIDHeader = "X-Request-ID"

// RequestLogger gives every request an ID and logs one line when it finishes:
//
//	12:00:01 INF request method=POST path=/api/v1/students status=201 latency=3.1 request_id=9f2c...
//
// It also stores a logger carrying that ID in the request's context, so
// anything logged while handling the request (zerolog.Ctx(ctx)) shows the
// same request_id. The ID goes back in the X-Request-ID response header too,
// which lets you match an error seen in the browser to its log line.
func RequestLogger(base zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		id := requestID(c.GetHeader(requestIDHeader))
		c.Header(requestIDHeader, id)
		log := base.With().Str("request_id", id).Logger()
		c.Request = c.Request.WithContext(log.WithContext(c.Request.Context()))

		c.Next() // run the rest of the chain: recovery, then the handler

		status := c.Writer.Status()
		event := log.Info()
		switch {
		case status >= 500:
			event = log.Error()
		case status >= 400:
			event = log.Warn()
		}
		event.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", status).
			Dur("latency", time.Since(start)). // milliseconds
			Int("bytes", c.Writer.Size()).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}

// Recoverer turns a panic in a handler into a 500 response and logs it, with
// the stack trace and request ID, instead of crashing the server. It must come
// after RequestLogger so the context already has the request's logger.
func Recoverer() gin.HandlerFunc {
	// A nil writer stops Gin from printing its own panic report; we log ours.
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		zerolog.Ctx(c.Request.Context()).Error().
			Interface("panic", recovered).
			Bytes("stack", debug.Stack()).
			Msg("panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

// A proxy in front of the API (like nginx) may already have given the request
// an ID; reuse it if it looks sane, so both logs share it. Otherwise make one.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func requestID(incoming string) string {
	if validRequestID.MatchString(incoming) {
		return incoming
	}
	b := make([]byte, 8)
	_, _ = io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}
