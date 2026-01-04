package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MetricsHandler returns current request metrics
func MetricsHandler(c *gin.Context) {
	metrics := GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"total_requests":       metrics.TotalRequests,
		"requests_by_endpoint": metrics.RequestsByEndpoint,
	})
}

