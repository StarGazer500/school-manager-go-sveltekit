package core

// Redis: connecting, and making the caches that server.go injects into the
// services (one rediscache.Cache per type, all with the same settings), plus
// the small helpers the services use with them.
//
// Redis is only a cache: everything in it can be rebuilt from Postgres. So the
// app keeps working when Redis is down; reads just go to the database.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/StarGazer500/rediscache"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/martin/school-api/internal/shared"
)

// NewCache makes a cache of T values whose keys all start with prefix, e.g.
// core.NewCache[students.Student](rdb, "student:", log). Redis errors the
// cache works around are logged to log, and its hits, misses and loads are
// reported as metrics (see cacheMetrics).
func NewCache[T any](rdb redis.UniversalClient, prefix string, log zerolog.Logger) *rediscache.Cache[T] {
	c := rediscache.New[T](rdb, options(prefix, log)...)
	if err := cacheMetrics(strings.TrimSuffix(prefix, ":"), c.Stats); err != nil {
		log.Warn().Err(err).Str("cache", prefix).Msg("cache metrics not registered")
	}
	return c
}

// cacheMetrics reports a cache's counters as three metrics, labelled with the
// cache's name ("student", "teacher"...). In Grafana, the hit rate is
//
//	app_cache_hits / (app_cache_hits + app_cache_misses)
//
// "Observable" instruments are read when metrics are exported (every 15s)
// instead of being updated on every call: stats already counts, so the
// callback just copies its numbers.
func cacheMetrics(name string, stats func() rediscache.Stats) error {
	meter := otel.Meter("github.com/martin/school-api/internal/core")
	hits, err := meter.Int64ObservableCounter("app.cache.hits",
		metric.WithDescription("Cache lookups answered from Redis (including remembered not-found)"))
	if err != nil {
		return err
	}
	misses, err := meter.Int64ObservableCounter("app.cache.misses",
		metric.WithDescription("Cache lookups not found in Redis"))
	if err != nil {
		return err
	}
	loads, err := meter.Int64ObservableCounter("app.cache.loads",
		metric.WithDescription("Times the cache had to load from the database"))
	if err != nil {
		return err
	}
	attrs := metric.WithAttributes(attribute.String("cache", name))
	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		s := stats()
		o.ObserveInt64(hits, int64(s.Hits), attrs)
		o.ObserveInt64(misses, int64(s.Misses), attrs)
		o.ObserveInt64(loads, int64(s.Loads), attrs)
		return nil
	}, hits, misses, loads)
	return err
}

// ToCache is for loaders passed to GetOrLoad: it turns the app's not-found
// error into rediscache.ErrNotFound, which tells the cache to remember that
// the record doesn't exist. Other errors pass through (and aren't cached).
func ToCache(err error) error {
	if errors.Is(err, shared.ErrNotFound) {
		return rediscache.ErrNotFound
	}
	return err
}

// FromCache does the opposite for GetOrLoad's result: rediscache.ErrNotFound
// becomes the app's not-found error with this message, so handlers answer 404.
func FromCache(err error, notFoundMessage string) error {
	if errors.Is(err, rediscache.ErrNotFound) {
		return shared.NotFound("%s", notFoundMessage)
	}
	return err
}

// OpenRedis creates a Redis client for url and checks the connection. Only a
// malformed URL is an error: if Redis isn't reachable, that is logged to log and
// the client is returned anyway, so the app starts and uses Postgres
// directly until Redis comes back.
//
// Every Redis command becomes a span in the request's trace (so a trace shows
// e.g. "GET student:5, 0.3ms"), and command latency and pool usage are
// reported as metrics.
func OpenRedis(ctx context.Context, url string, log zerolog.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("REDIS_URL: %w", err)
	}
	// Fail fast. go-redis retries a lot by default (3 retries per command, 5
	// connection attempts each), which made every request take seconds while
	// Redis was down. For a cache, giving up quickly and reading Postgres is
	// better than waiting.
	opts.MaxRetries = -1 // -1 means no retries (0 means "use the default")
	opts.DialerRetries = 1
	opts.DialTimeout = 300 * time.Millisecond
	opts.ReadTimeout = 300 * time.Millisecond
	opts.WriteTimeout = 300 * time.Millisecond

	// go-redis has its own logger, which prints with Go's standard log package
	// and would bypass our logger (and break JSON output). Route it through ours.
	// go-redis only offers this setting for the whole process, not per client.
	redis.SetLogger(redisLogger{log: log})

	rdb := redis.NewClient(opts)
	// WithDBStatement(false): spans name the command (GET, SET...) but don't
	// record its arguments. Those include the cached values, i.e. students'
	// names and emails, which don't belong in a tracing system.
	if err := redisotel.InstrumentTracing(rdb, redisotel.WithDBStatement(false)); err != nil {
		return nil, fmt.Errorf("redis tracing: %w", err)
	}
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return nil, fmt.Errorf("redis metrics: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Warn().Err(err).Str("addr", opts.Addr).
			Msg("redis unavailable; continuing without a cache (try `make db-up`)")
	}
	return rdb, nil
}

// redisLogger adapts our logger to the interface go-redis logs through. Its
// messages (e.g. "failed to dial") repeat what our own warnings already say,
// so they're logged at debug level: visible with LOG_LEVEL=debug only.
type redisLogger struct {
	log zerolog.Logger
}

func (l redisLogger) Printf(ctx context.Context, format string, v ...any) {
	l.log.Debug().Ctx(ctx).Str("source", "go-redis").Msgf(format, v...)
}

// options returns the settings every cache in the app uses, plus its own key
// prefix (e.g. "student:"):
//   - entries expire after 10 minutes, minus up to 10% so they don't all expire together
//   - "not found" answers are remembered for 1 minute
//   - Redis errors that GetOrLoad works around (it falls back to the
//     database) are logged to log as warnings
func options(prefix string, log zerolog.Logger) []rediscache.Option {
	return []rediscache.Option{
		rediscache.WithPrefix(prefix),
		rediscache.WithTTL(10 * time.Minute),
		rediscache.WithJitter(0.1),
		rediscache.WithNegativeTTL(time.Minute),
		rediscache.WithErrorHandler(func(err error) {
			log.Warn().Err(err).Str("cache", prefix).Msg("cache error, used the database instead")
		}),
	}
}

// WarnInvalidate logs a failed invalidation, with the request's logger from
// ctx (it's only called while handling a request). The database write already
// succeeded, so the request shouldn't fail, but the old entry may be served
// until it expires.
func WarnInvalidate(ctx context.Context, err error, what string) {
	if err != nil {
		zerolog.Ctx(ctx).Warn().Ctx(ctx).Err(err).Str("entry", what).
			Msg("couldn't remove cache entry; it may be stale until it expires")
	}
}
