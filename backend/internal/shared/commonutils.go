// Package shared holds the code every feature uses:
//   - commonutils.go: the error kinds services return (NotFound, Conflict,
//     Invalid) and the Gin helpers handlers use (ErrorResponse, ValidationMessage, IDParam, ...)
package shared

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ---------------------------------------------------------------------------
// Error kinds

// The error kinds services return. Handlers turn them into HTTP status codes
// (see ErrorResponse below), so services never need to know about HTTP.
// Check them with errors.Is(err, shared.ErrNotFound).
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid input")
)

// AppError carries a user-facing message plus one of the kinds above.
type AppError struct {
	Kind    error
	Message string
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Kind }

func NotFound(format string, args ...any) error {
	return &AppError{Kind: ErrNotFound, Message: fmt.Sprintf(format, args...)}
}

func Conflict(format string, args ...any) error {
	return &AppError{Kind: ErrConflict, Message: fmt.Sprintf(format, args...)}
}

func Invalid(format string, args ...any) error {
	return &AppError{Kind: ErrInvalid, Message: fmt.Sprintf(format, args...)}
}

// ---------------------------------------------------------------------------
// HTTP helpers for handlers

// ErrorResponse works out what to send back for err: the HTTP status for its
// kind and a message that is safe to show. Unknown errors become 500 with a
// generic message, and the real error is logged, so internals never leak to
// clients. The handler writes the response:
//
//	status, msg := shared.ErrorResponse(c, err)
//	c.AbortWithStatusJSON(status, gin.H{"error": msg})
func ErrorResponse(c *gin.Context, err error) (status int, msg string) {
	status = http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, ErrInvalid):
		status = http.StatusBadRequest
	}

	msg = "internal server error"
	var appErr *AppError
	if errors.As(err, &appErr) {
		msg = appErr.Message
	} else {
		// The real error is logged and recorded on the request's span (so the
		// trace in Grafana shows it), but never sent to the client. The client
		// gets the trace ID in the X-Trace-ID header to report instead.
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zerolog.Ctx(ctx).Error().Err(err).Str("route", c.FullPath()).Msg("unhandled error")
	}
	return status, msg
}

// ValidationMessage turns an error from c.ShouldBindJSON into a readable
// message: failed `binding:"..."` rules become e.g. "email must be a valid
// email", and anything else (malformed JSON) becomes "invalid JSON body".
// The handler sends it as a 400:
//
//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
func ValidationMessage(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return "invalid JSON body"
	}
	msgs := make([]string, len(ve))
	for i, fe := range ve {
		msgs[i] = describe(fe)
	}
	return strings.Join(msgs, "; ")
}

func describe(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email"
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	default:
		return fe.Field() + " is invalid"
	}
}

// UseJSONFieldNames makes validation errors report "firstName" (the json tag)
// instead of the Go field name "FirstName". Call once at startup.
func UseJSONFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
}

// IDParam parses a positive integer path parameter such as /students/:id.
func IDParam(c *gin.Context, name string) (int64, error) {
	return parseID(c.Param(name), name)
}

// OptionalIDQuery parses ?name=123, returning 0 when the parameter is absent.
func OptionalIDQuery(c *gin.Context, name string) (int64, error) {
	raw := c.Query(name)
	if raw == "" {
		return 0, nil
	}
	return parseID(raw, name)
}

func parseID(raw, name string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, Invalid("%s must be a positive integer", name)
	}
	return id, nil
}
