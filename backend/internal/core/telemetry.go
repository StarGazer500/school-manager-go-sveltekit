package core

// OpenTelemetry: the app's traces and metrics.
//
//   - Traces show where one request spent its time: the HTTP handler, each
//     SQL query, each Redis call, as a timeline. They're pushed over OTLP to
//     the OpenTelemetry Collector (observability/), which decides which ones
//     to keep (tail sampling) and forwards them to Tempo.
//   - Metrics are numbers over time: request rate, errors, latency, pool
//     usage, cache hits, Go memory... They're *pulled*: the app serves them on
//     /metrics (port METRICS_PORT, not the public API port) and Prometheus
//     scrapes it every 15 seconds. Because Prometheus does the asking, it
//     knows when the app stops answering (the `up` metric).
//   - Logs aren't handled here: the app writes them to stdout and a file, and
//     the collector ships the file to Loki (see internal/logger).
//
// The providers are registered globally (otel.SetTracerProvider and friends).
// That's OpenTelemetry's design: instrumentation libraries such as the Gin,
// pgx and Redis ones find them there, so everything shares one pipeline.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Telemetry holds what main.go needs after setup.
type Telemetry struct {
	// MetricsHandler serves the metrics for Prometheus to scrape. main.go
	// serves it on its own port, so /metrics isn't exposed with the API.
	MetricsHandler http.Handler

	shutdowns []func(context.Context) error
}

// Shutdown flushes the spans still buffered and stops the exporters. Call it
// last on the way out, or the final seconds of traces are lost.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, fn := range t.shutdowns {
		errs = append(errs, fn(ctx))
	}
	return errors.Join(errs...)
}

// SetupTelemetry starts the trace and metric pipelines. If cfg turns
// telemetry off (OTEL_SDK_DISABLED=true), it does nothing: OpenTelemetry's
// default global providers are no-ops, so the instrumented code still runs,
// and MetricsHandler answers 404.
//
// Nothing here connects to the collector yet. If it isn't running, span
// exports fail in the background and are reported through the error handler
// (ReportTelemetryErrors); the app itself carries on.
func SetupTelemetry(ctx context.Context, cfg Config) (*Telemetry, error) {
	t := &Telemetry{MetricsHandler: http.NotFoundHandler()}
	if !cfg.TelemetryEnabled {
		return t, nil
	}

	// The resource describes who is sending: shown on every trace and metric,
	// and how Grafana tells services, versions and environments apart.
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("telemetry resource: %w", err)
	}

	// Traces. Every span is sent; the collector decides which traces to keep
	// (all errors and slow ones, a share of the rest). It can only do that
	// if it sees whole traces, so the app doesn't sample on its own (the SDK
	// default, which the standard OTEL_TRACES_SAMPLER variable can override).
	//
	// The batcher buffers spans in memory (2048 by default, OTEL_BSP_MAX_QUEUE_SIZE)
	// and exports them in the background. If the buffer fills up under heavy
	// traffic, new spans are dropped rather than slowing requests down.
	traceExp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OTLPEndpoint+"/v1/traces"))
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	t.shutdowns = append(t.shutdowns, tp.Shutdown)

	// Propagation: read and write the W3C `traceparent` header, so a trace
	// started by a caller (the browser, through Faro) continues here.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Metrics, served for Prometheus to scrape. The service's name, version
	// and environment become labels on every series, so dashboards can
	// filter by them. A registry of our own (instead of Prometheus's global
	// one) keeps /metrics to exactly what we report.
	registry := promclient.NewRegistry()
	metricExp, err := otelprom.New(
		otelprom.WithRegisterer(registry),
		otelprom.WithResourceAsConstantLabels(attribute.NewAllowKeysFilter(
			semconv.ServiceNameKey, semconv.ServiceVersionKey, semconv.DeploymentEnvironmentNameKey,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricExp), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)
	t.shutdowns = append(t.shutdowns, mp.Shutdown)
	// OpenMetrics format carries exemplars: a sample trace ID attached to
	// latency data points, which Grafana shows as clickable dots.
	t.MetricsHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true})

	// Go runtime metrics: memory, goroutines, garbage collection.
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(15 * time.Second)); err != nil {
		return nil, fmt.Errorf("runtime metrics: %w", err)
	}
	return t, nil
}

// ReportTelemetryErrors sends OpenTelemetry's own errors (typically "can't
// reach the collector") to log, at most one a minute, so a stopped collector
// gives a hint instead of either silence or a flood.
func ReportTelemetryErrors(log zerolog.Logger) {
	var mu sync.Mutex
	var last time.Time
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if time.Since(last) < time.Minute {
			return
		}
		last = time.Now()
		log.Warn().Err(err).
			Msg("telemetry export failed (is the observability stack running? `make obs-up`); further errors muted for a minute")
	}))
}
