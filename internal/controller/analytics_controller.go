package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"irrigation-analytics/internal/service"

	"github.com/gin-gonic/gin"
)

// AnalyticsController handles analytics-related HTTP requests
type AnalyticsController struct {
	analyticsService service.AnalyticsService
	logger           *slog.Logger
}

// NewAnalyticsController creates a new analytics controller
func NewAnalyticsController(analyticsService service.AnalyticsService, logger *slog.Logger) *AnalyticsController {
	return &AnalyticsController{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// GetIrrigationAnalytics handles GET /v1/farms/{farm_id}/irrigation/analytics
// Query parameters:
//   - sector_id (optional): Filter by sector ID
//   - start_date (required): Start date in ISO 8601 format (RFC3339 or YYYY-MM-DD)
//   - end_date (required): End date in ISO 8601 format (RFC3339 or YYYY-MM-DD)
//   - aggregation (optional): daily, weekly, or monthly (default: daily)
func (c *AnalyticsController) GetIrrigationAnalytics(ctx *gin.Context) {
	startTime := time.Now()
	// Parse farm_id from path
	farmIDStr := ctx.Param("farm_id")
	farmID, err := strconv.ParseUint(farmIDStr, 10, 32)
	if err != nil {
		c.logger.Warn("invalid farm_id",
			"farm_id", farmIDStr,
			"error", err.Error(),
		)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid farm_id",
			"message": "farm_id must be a valid unsigned integer",
		})
		return
	}

	// Parse optional sector_id from query
	var sectorID *uint
	if sectorIDStr := ctx.Query("sector_id"); sectorIDStr != "" {
		sid, err := strconv.ParseUint(sectorIDStr, 10, 32)
		if err != nil {
			c.logger.Warn("invalid sector_id",
				"sector_id", sectorIDStr,
				"farm_id", farmID,
				"error", err.Error(),
			)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid sector_id",
				"message": "sector_id must be a valid unsigned integer",
			})
			return
		}
		sidUint := uint(sid)
		sectorID = &sidUint
	}

	// Parse start_date from query
	startDateStr := ctx.Query("start_date")
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameter",
			"message": "start_date is required",
		})
		return
	}

	startDate, err := parseISO8601Date(startDateStr)
	if err != nil {
		c.logger.Warn("invalid start_date",
			"start_date", startDateStr,
			"farm_id", farmID,
			"error", err.Error(),
		)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_date",
			"message": "start_date must be in ISO 8601 format (RFC3339 or YYYY-MM-DD)",
		})
		return
	}

	// Parse end_date from query
	endDateStr := ctx.Query("end_date")
	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameter",
			"message": "end_date is required",
		})
		return
	}

	endDate, err := parseISO8601Date(endDateStr)
	if err != nil {
		c.logger.Warn("invalid end_date",
			"end_date", endDateStr,
			"farm_id", farmID,
			"error", err.Error(),
		)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_date",
			"message": "end_date must be in ISO 8601 format (RFC3339 or YYYY-MM-DD)",
		})
		return
	}

	// Validate date range
	if endDate.Before(startDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid date range",
			"message": "end_date must be after start_date",
		})
		return
	}

	// Parse aggregation level (optional, default: daily)
	aggregation := ctx.DefaultQuery("aggregation", "daily")
	if aggregation != "daily" && aggregation != "weekly" && aggregation != "monthly" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid aggregation",
			"message": "aggregation must be one of: daily, weekly, monthly",
		})
		return
	}

	// Check if farm exists
	farmExists, err := c.analyticsService.FarmExists(uint(farmID))
	if err != nil {
		latency := time.Since(startTime)
		c.logger.Error("failed to check farm existence",
			"farm_id", farmID,
			"error", err.Error(),
			"latency_ms", latency.Milliseconds(),
		)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to verify farm existence",
		})
		return
	}
	if !farmExists {
		latency := time.Since(startTime)
		c.logger.Warn("farm not found",
			"farm_id", farmID,
			"latency_ms", latency.Milliseconds(),
		)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "Farm not found",
			"message": fmt.Sprintf("Farm with ID %d does not exist", farmID),
		})
		return
	}

	// Log query parameters
	c.logger.Info("processing analytics request",
		"farm_id", farmID,
		"sector_id", sectorID,
		"start_date", startDate.Format(time.RFC3339),
		"end_date", endDate.Format(time.RFC3339),
		"aggregation", aggregation,
	)

	// Call service
	analytics, err := c.analyticsService.GetIrrigationAnalytics(
		uint(farmID),
		sectorID,
		startDate,
		endDate,
		aggregation,
	)
	if err != nil {
		latency := time.Since(startTime)
		c.logger.Error("failed to retrieve analytics",
			"farm_id", farmID,
			"sector_id", sectorID,
			"start_date", startDate.Format(time.RFC3339),
			"end_date", endDate.Format(time.RFC3339),
			"aggregation", aggregation,
			"error", err.Error(),
			"latency_ms", latency.Milliseconds(),
		)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve analytics data",
		})
		return
	}

	latency := time.Since(startTime)
	c.logger.Info("analytics request completed",
		"farm_id", farmID,
		"sector_id", sectorID,
		"aggregation", aggregation,
		"data_points", len(analytics.Data),
		"latency_ms", latency.Milliseconds(),
	)

	ctx.JSON(http.StatusOK, analytics)
}

// parseISO8601Date parses a date string in ISO 8601 format (RFC3339 is ISO 8601 compliant)
// Supports:
//   - RFC3339 (e.g., "2006-01-02T15:04:05Z07:00")
//   - RFC3339Nano (e.g., "2006-01-02T15:04:05.999999999Z07:00")
//   - YYYY-MM-DD (e.g., "2006-01-02")
//   - YYYY-MM-DDTHH:MM:SS (e.g., "2006-01-02T15:04:05")
//   - YYYY-MM-DDTHH:MM:SSZ (e.g., "2006-01-02T15:04:05Z")
func parseISO8601Date(dateStr string) (time.Time, error) {
	// Try RFC3339 format first (ISO 8601 compliant)
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// Try RFC3339Nano format
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t, nil
	}

	// Try YYYY-MM-DD format (ISO 8601 date format)
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		// Set to start of day in UTC
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	// Try YYYY-MM-DDTHH:MM:SS format (ISO 8601 without timezone)
	if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
		return t, nil
	}

	// Try YYYY-MM-DDTHH:MM:SSZ format (ISO 8601 with Z timezone)
	if t, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
		return t, nil
	}

	// Try YYYY-MM-DDTHH:MM:SS+HH:MM format (ISO 8601 with timezone offset)
	if t, err := time.Parse("2006-01-02T15:04:05Z07:00", dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse ISO 8601 date: %s (expected RFC3339 or YYYY-MM-DD format)", dateStr)
}
