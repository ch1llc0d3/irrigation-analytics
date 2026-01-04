package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"irrigation-analytics/internal/service"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// mockAnalyticsService is a mock implementation of AnalyticsService for testing
type mockAnalyticsService struct {
	analytics *service.AnalyticsResponse
	err       error
}

func (m *mockAnalyticsService) GetIrrigationAnalytics(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string) (*service.AnalyticsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.analytics, nil
}

func setupRouter(controller *AnalyticsController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/v1")
	{
		farms := v1.Group("/farms")
		{
			farms.GET("/:farm_id/irrigation/analytics", controller.GetIrrigationAnalytics)
		}
	}
	return r
}

func TestGetIrrigationAnalytics_Success(t *testing.T) {
	// Create mock service
	mockService := &mockAnalyticsService{
		analytics: &service.AnalyticsResponse{
			FarmID:      1,
			SectorID:    nil,
			Aggregation: "daily",
			Period: service.PeriodInfo{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			},
			Data: []service.AggregatedDataPoint{
				{
					Period:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					WaterVolume: 100.0,
					Duration:    100,
					Efficiency:  1.0,
					EventCount:  1,
				},
			},
			Summary: service.AnalyticsSummary{
				TotalWaterVolume:  100.0,
				TotalDuration:     100,
				AverageEfficiency: 1.0,
				TotalEvents:       1,
			},
			YearOverYear: service.YearOverYearComparison{},
		},
	}

	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	// Create request
	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&aggregation=daily", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response service.AnalyticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.FarmID != 1 {
		t.Errorf("Expected FarmID 1, got %d", response.FarmID)
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(response.Data))
	}
}

func TestGetIrrigationAnalytics_InvalidFarmID(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/invalid/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var errorResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResponse["error"] != "Invalid farm_id" {
		t.Errorf("Expected error 'Invalid farm_id', got %v", errorResponse["error"])
	}
}

func TestGetIrrigationAnalytics_MissingStartDate(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?end_date=2024-01-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetIrrigationAnalytics_MissingEndDate(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetIrrigationAnalytics_InvalidDateRange(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-31&end_date=2024-01-01", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetIrrigationAnalytics_InvalidAggregation(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&aggregation=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetIrrigationAnalytics_WithSectorID(t *testing.T) {
	mockService := &mockAnalyticsService{
		analytics: &service.AnalyticsResponse{
			FarmID:      1,
			SectorID:    uintPtr(2),
			Aggregation: "daily",
			Period: service.PeriodInfo{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			},
			Data:    []service.AggregatedDataPoint{},
			Summary: service.AnalyticsSummary{},
		},
	}

	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&sector_id=2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response service.AnalyticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.SectorID == nil || *response.SectorID != 2 {
		t.Errorf("Expected SectorID 2, got %v", response.SectorID)
	}
}

func TestGetIrrigationAnalytics_InvalidSectorID(t *testing.T) {
	mockService := &mockAnalyticsService{}
	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&sector_id=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetIrrigationAnalytics_ISO8601DateFormats(t *testing.T) {
	mockService := &mockAnalyticsService{
		analytics: &service.AnalyticsResponse{
			FarmID:      1,
			Aggregation: "daily",
			Data:        []service.AggregatedDataPoint{},
			Summary:     service.AnalyticsSummary{},
		},
	}

	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	// Test RFC3339 format
	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("RFC3339 format failed: expected %d, got %d", http.StatusOK, w.Code)
	}

	// Test YYYY-MM-DD format
	req, _ = http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("YYYY-MM-DD format failed: expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetIrrigationAnalytics_ServiceError(t *testing.T) {
	mockService := &mockAnalyticsService{
		err: &serviceError{message: "database connection failed"},
	}

	logger := slog.Default()
	controller := NewAnalyticsController(mockService, logger)
	router := setupRouter(controller)

	req, _ := http.NewRequest("GET", "/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// serviceError is a simple error type for testing
type serviceError struct {
	message string
}

func (e *serviceError) Error() string {
	return e.message
}

// Helper function
func uintPtr(u uint) *uint {
	return &u
}

