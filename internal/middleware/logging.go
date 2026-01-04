package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestMetrics holds in-memory request metrics
type RequestMetrics struct {
	mu            sync.RWMutex
	TotalRequests uint64
	RequestsByEndpoint map[string]uint64
}

var metrics = &RequestMetrics{
	RequestsByEndpoint: make(map[string]uint64),
}

// GetMetrics returns the current request metrics
func GetMetrics() RequestMetrics {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return RequestMetrics{
		TotalRequests:      metrics.TotalRequests,
		RequestsByEndpoint: copyMap(metrics.RequestsByEndpoint),
	}
}

// copyMap creates a copy of the map
func copyMap(src map[string]uint64) map[string]uint64 {
	dst := make(map[string]uint64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// StructuredLoggingMiddleware provides structured logging with request latency and query parameters
func StructuredLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log request start with query parameters
		logger.Info("request started",
			"method", method,
			"path", path,
			"query_params", c.Request.URL.Query().Encode(),
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Update metrics
		metrics.mu.Lock()
		metrics.TotalRequests++
		endpoint := method + " " + path
		metrics.RequestsByEndpoint[endpoint]++
		metrics.mu.Unlock()

		// Log request completion
		logger.Info("request completed",
			"method", method,
			"path", path,
			"status_code", statusCode,
			"latency_ms", latency.Milliseconds(),
			"latency", latency.String(),
			"bytes_written", c.Writer.Size(),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error",
					"method", method,
					"path", path,
					"error", err.Error(),
					"latency_ms", latency.Milliseconds(),
				)
			}
		}
	}
}

