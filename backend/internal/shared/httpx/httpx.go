// Package httpx holds small Gin helpers shared by every feature's handler.
package httpx

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/martin/school-api/internal/shared/apperr"
)

// Error writes err as a JSON response with a status code matching its kind.
// Unknown errors become 500 and are logged, so internals never leak to clients.
func Error(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, apperr.ErrInvalid):
		status = http.StatusBadRequest
	}

	msg := "internal server error"
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		msg = appErr.Message
	} else {
		log.Printf("unhandled error on %s %s: %v", c.Request.Method, c.FullPath(), err)
	}

	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}

// BadRequest is used for request binding/parsing failures. Validation errors
// from `binding:"..."` tags are turned into readable messages.
func BadRequest(c *gin.Context, err error) {
	msg := "invalid JSON body"
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, len(ve))
		for i, fe := range ve {
			msgs[i] = describe(fe)
		}
		msg = strings.Join(msgs, "; ")
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
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
		return 0, apperr.Invalid("%s must be a positive integer", name)
	}
	return id, nil
}
