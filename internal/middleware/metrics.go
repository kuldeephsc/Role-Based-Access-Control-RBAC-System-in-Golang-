package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"rbac-platform/internal/platform/metrics"
)

// HTTPMetrics records request duration and status for every route, backing
// the rbac_http_request_duration_seconds histogram scraped by Prometheus.
func HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(duration)
	}
}
