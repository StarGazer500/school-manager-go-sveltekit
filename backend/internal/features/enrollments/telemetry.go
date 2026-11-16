package enrollments

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Business telemetry for enrollments: a span around Enroll, so a trace shows
// the business step between the HTTP request and its SQL queries, and two
// counters that answer "how many enrollments, and why were some refused?".
//
// OpenTelemetry's convention is one tracer and meter per package, named
// after it, taken from the global providers that core.SetupTelemetry
// registers. If telemetry is off, they do nothing.

const scope = "github.com/martin/school-api/internal/features/enrollments"

var tracer = otel.Tracer(scope)

type metrics struct {
	created  metric.Int64Counter
	rejected metric.Int64Counter
}

func newMetrics() metrics {
	meter := otel.Meter(scope)
	// Creating an instrument only fails for an invalid name, which these
	// aren't; a failed one is a no-op, so the error can be ignored.
	created, _ := meter.Int64Counter("app.enrollments.created",
		metric.WithDescription("Students enrolled in a course"))
	rejected, _ := meter.Int64Counter("app.enrollments.rejected",
		metric.WithDescription("Enrollments refused, by reason"))
	return metrics{created: created, rejected: rejected}
}

// Why Enroll can refuse. They label the rejected counter, so keep the set
// small and fixed (never put ids or free text in metric labels).
const (
	reasonCourseFull      = "course_full"
	reasonCourseNotFound  = "course_not_found"
	reasonAlreadyEnrolled = "already_enrolled"
	reasonStudentNotFound = "student_not_found"
)

// reject records why an enrollment was refused, on the span and as a metric.
// Refusals are expected outcomes (they answer 4xx), so the span isn't marked
// as an error; unexpected errors are recorded by the HTTP layer.
func (m metrics) reject(ctx context.Context, reason string) {
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("enrollment.rejected", reason))
	m.rejected.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
}
