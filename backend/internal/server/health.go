package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Health checks, the standard pair that Docker, Kubernetes and load balancers
// call every few seconds. They're kept out of traces and request logs (see
// isHealthCheck in middleware.go) so they don't bury real traffic.

// liveness answers GET /healthz: "is the process running and serving HTTP?"
// It checks nothing else on purpose. If it fails, the process is stuck and
// the platform should restart it; a database outage is no reason to restart.
func liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readiness answers GET /readyz: "can this instance serve requests right
// now?" It needs Postgres, so if Postgres doesn't answer it returns 503, and
// a load balancer stops sending traffic here until it recovers. Redis is only
// a cache, so Redis being down is reported but still counts as ready.
func readiness(db *pgxpool.Pool, rdb redis.UniversalClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		checks := gin.H{"postgres": "ok", "redis": "ok"}
		status, overall := http.StatusOK, "ready"

		if err := db.Ping(ctx); err != nil {
			checks["postgres"] = err.Error()
			status, overall = http.StatusServiceUnavailable, "not ready"
		}
		if err := rdb.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			if status == http.StatusOK {
				overall = "degraded: running without the cache"
			}
		}
		c.JSON(status, gin.H{"status": overall, "checks": checks})
	}
}
