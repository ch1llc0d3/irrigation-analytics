package service

import (
	"math"
	"time"

	"irrigation-analytics/internal/repository"
)

// AnalyticsService defines the interface for analytics operations
type AnalyticsService interface {
	FarmExists(farmID uint) (bool, error)
	GetIrrigationAnalytics(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string) (*AnalyticsResponse, error)
}

// AnalyticsResponse represents the analytics data response
type AnalyticsResponse struct {
	FarmID           uint                   `json:"farm_id"`
	SectorID         *uint                  `json:"sector_id,omitempty"`
	Period           PeriodInfo             `json:"period"`
	Aggregation      string                 `json:"aggregation"`
	Data             []AggregatedDataPoint  `json:"data"`
	Summary          AnalyticsSummary       `json:"summary"`
	PeriodComparison PeriodComparison       `json:"period_comparison"`
	SectorBreakdown  []SectorBreakdown      `json:"sector_breakdown,omitempty"`
	YearOverYear     YearOverYearComparison `json:"year_over_year"`
}

// PeriodInfo contains date range information
type PeriodInfo struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// AggregatedDataPoint represents a single aggregated data point
type AggregatedDataPoint struct {
	Period        time.Time `json:"period"`
	WaterVolume   float64   `json:"water_volume"`
	Duration      int       `json:"duration"`   // in minutes
	Efficiency    float64   `json:"efficiency"` // real_amount / nominal_amount
	EventCount    int       `json:"event_count"`
	RealAmount    float64   `json:"real_amount"`
	NominalAmount float64   `json:"nominal_amount"`
}

// AnalyticsSummary contains summary statistics
type AnalyticsSummary struct {
	TotalWaterVolume   float64 `json:"total_water_volume"`
	TotalDuration      int     `json:"total_duration"` // in minutes
	AverageEfficiency  float64 `json:"average_efficiency"`
	TotalEvents        int     `json:"total_events"`
	TotalRealAmount    float64 `json:"total_real_amount"`
	TotalNominalAmount float64 `json:"total_nominal_amount"`
}

// PeriodComparison contains comparison metrics between periods
type PeriodComparison struct {
	OneYearAgo  *PeriodMetrics `json:"one_year_ago,omitempty"`
	TwoYearsAgo *PeriodMetrics `json:"two_years_ago,omitempty"`
}

// PeriodMetrics contains metrics for a specific period with percentage changes
type PeriodMetrics struct {
	Period                  PeriodInfo `json:"period"`
	TotalWaterVolume        float64    `json:"total_water_volume"`
	TotalEvents             int        `json:"total_events"`
	AverageEfficiency       float64    `json:"average_efficiency"`
	VolumeChangePercent     float64    `json:"volume_change_percent"`
	EventsChangePercent     float64    `json:"events_change_percent"`
	EfficiencyChangePercent float64    `json:"efficiency_change_percent"`
}

// SectorBreakdown contains analytics broken down by sector
type SectorBreakdown struct {
	SectorID           uint    `json:"sector_id"`
	TotalWaterVolume   float64 `json:"total_water_volume"`
	TotalEvents        int     `json:"total_events"`
	AverageEfficiency  float64 `json:"average_efficiency"`
	TotalRealAmount    float64 `json:"total_real_amount"`
	TotalNominalAmount float64 `json:"total_nominal_amount"`
}

// YearOverYearComparison contains YoY comparison data
type YearOverYearComparison struct {
	OneYearAgo  *YearComparison `json:"one_year_ago,omitempty"`
	TwoYearsAgo *YearComparison `json:"two_years_ago,omitempty"`
}

// YearComparison contains comparison metrics for a specific year
type YearComparison struct {
	Period            PeriodInfo `json:"period"`
	TotalWaterVolume  float64    `json:"total_water_volume"`
	TotalDuration     int        `json:"total_duration"`
	AverageEfficiency float64    `json:"average_efficiency"`
	TotalEvents       int        `json:"total_events"`
	ChangePercent     float64    `json:"change_percent"` // Percentage change from current period
}

// analyticsService implements AnalyticsService
type analyticsService struct {
	repo repository.IrrigationRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repo repository.IrrigationRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

// FarmExists checks if a farm exists
func (s *analyticsService) FarmExists(farmID uint) (bool, error) {
	return s.repo.FarmExists(farmID)
}

// GetIrrigationAnalytics retrieves and processes irrigation analytics
func (s *analyticsService) GetIrrigationAnalytics(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string) (*AnalyticsResponse, error) {
	// Validate aggregation level
	if aggregation == "" {
		aggregation = "daily"
	}
	if aggregation != "daily" && aggregation != "weekly" && aggregation != "monthly" {
		aggregation = "daily"
	}

	// Fetch current period data
	currentData, err := s.repo.GetAggregatedData(farmID, sectorID, startDate, endDate, aggregation)
	if err != nil {
		return nil, err
	}

	// Process current period data
	dataPoints := s.processDataPoints(currentData, aggregation)
	summary := s.calculateSummary(currentData)

	// Calculate period comparison (YoY with detailed metrics)
	periodComparison := s.calculatePeriodComparison(farmID, sectorID, startDate, endDate, aggregation, summary)

	// Calculate sector breakdown (if not filtering by specific sector)
	var sectorBreakdown []SectorBreakdown
	if sectorID == nil {
		sectorBreakdown = s.calculateSectorBreakdown(farmID, startDate, endDate, aggregation)
	}

	// Fetch YoY data (legacy format for backward compatibility)
	yoy := s.calculateYearOverYear(farmID, sectorID, startDate, endDate, aggregation, summary)

	return &AnalyticsResponse{
		FarmID:   farmID,
		SectorID: sectorID,
		Period: PeriodInfo{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Aggregation:      aggregation,
		Data:             dataPoints,
		Summary:          summary,
		PeriodComparison: periodComparison,
		SectorBreakdown:  sectorBreakdown,
		YearOverYear:     yoy,
	}, nil
}

// calculateEfficiency calculates efficiency = real_amount / nominal_amount
// Handles division by zero gracefully
func (s *analyticsService) calculateEfficiency(realAmount, nominalAmount float64) float64 {
	if nominalAmount == 0 {
		// If nominal amount is 0, return 0 efficiency to avoid division by zero
		// This handles edge cases where nominal amount hasn't been set
		return 0.0
	}
	efficiency := realAmount / nominalAmount
	return math.Round(efficiency*10000) / 10000 // Round to 4 decimal places
}

// processDataPoints converts raw data to aggregated data points with efficiency calculation
func (s *analyticsService) processDataPoints(data []repository.AggregatedDataWithCount, aggregation string) []AggregatedDataPoint {
	points := make([]AggregatedDataPoint, 0, len(data))

	for _, item := range data {
		d := item.Data
		// Calculate efficiency using RealAmount and NominalAmount
		efficiency := s.calculateEfficiency(d.RealAmount, d.NominalAmount)

		// If RealAmount or NominalAmount are not set, fall back to water_volume calculation
		if d.RealAmount == 0 && d.NominalAmount == 0 && d.WaterVolume > 0 {
			// Fallback: use water_volume as real and calculate nominal from duration
			if d.Duration > 0 {
				nominalVolume := float64(d.Duration) * 1.0 // 1 liter per minute
				efficiency = s.calculateEfficiency(d.WaterVolume, nominalVolume)
			}
		}

		points = append(points, AggregatedDataPoint{
			Period:        d.StartTime,
			WaterVolume:   d.WaterVolume,
			Duration:      d.Duration,
			Efficiency:    efficiency,
			EventCount:    item.EventCount, // Use event_count from aggregation
			RealAmount:    d.RealAmount,
			NominalAmount: d.NominalAmount,
		})
	}

	return points
}

// calculateSummary computes summary statistics
func (s *analyticsService) calculateSummary(data []repository.AggregatedDataWithCount) AnalyticsSummary {
	var totalWaterVolume float64
	var totalDuration int
	var totalEfficiency float64
	var efficiencyCount int
	var totalRealAmount float64
	var totalNominalAmount float64
	var totalEvents int

	for _, item := range data {
		d := item.Data
		totalWaterVolume += d.WaterVolume
		totalDuration += d.Duration
		totalRealAmount += d.RealAmount
		totalNominalAmount += d.NominalAmount
		totalEvents += item.EventCount // Sum event counts from aggregation

		// Calculate efficiency for summary
		efficiency := s.calculateEfficiency(d.RealAmount, d.NominalAmount)

		// If efficiency couldn't be calculated from RealAmount/NominalAmount, use fallback
		if efficiency == 0 && d.WaterVolume > 0 && d.Duration > 0 {
			nominalVolume := float64(d.Duration) * 1.0
			efficiency = s.calculateEfficiency(d.WaterVolume, nominalVolume)
		}

		if efficiency > 0 {
			totalEfficiency += efficiency
			efficiencyCount++
		}
	}

	avgEfficiency := 0.0
	if efficiencyCount > 0 {
		avgEfficiency = totalEfficiency / float64(efficiencyCount)
	}

	return AnalyticsSummary{
		TotalWaterVolume:   math.Round(totalWaterVolume*100) / 100,
		TotalDuration:      totalDuration,
		AverageEfficiency:  math.Round(avgEfficiency*10000) / 10000,
		TotalEvents:        totalEvents,
		TotalRealAmount:    math.Round(totalRealAmount*100) / 100,
		TotalNominalAmount: math.Round(totalNominalAmount*100) / 100,
	}
}

// calculatePeriodComparison computes period comparison with percentage changes for volume, events, and efficiency
func (s *analyticsService) calculatePeriodComparison(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string, currentSummary AnalyticsSummary) PeriodComparison {
	comparison := PeriodComparison{}

	// Fetch data for -1 year
	oneYearData, err := s.repo.GetYearOverYearData(farmID, sectorID, startDate, endDate, aggregation, 1)
	if err == nil && len(oneYearData) > 0 {
		oneYearSummary := s.calculateSummary(oneYearData)

		comparison.OneYearAgo = &PeriodMetrics{
			Period: PeriodInfo{
				StartDate: startDate.AddDate(-1, 0, 0),
				EndDate:   endDate.AddDate(-1, 0, 0),
			},
			TotalWaterVolume:        oneYearSummary.TotalWaterVolume,
			TotalEvents:             oneYearSummary.TotalEvents,
			AverageEfficiency:       oneYearSummary.AverageEfficiency,
			VolumeChangePercent:     s.calculateChangePercent(currentSummary.TotalWaterVolume, oneYearSummary.TotalWaterVolume),
			EventsChangePercent:     s.calculateChangePercent(float64(currentSummary.TotalEvents), float64(oneYearSummary.TotalEvents)),
			EfficiencyChangePercent: s.calculateChangePercent(currentSummary.AverageEfficiency, oneYearSummary.AverageEfficiency),
		}
	}

	// Fetch data for -2 years
	twoYearsData, err := s.repo.GetYearOverYearData(farmID, sectorID, startDate, endDate, aggregation, 2)
	if err == nil && len(twoYearsData) > 0 {
		twoYearsSummary := s.calculateSummary(twoYearsData)

		comparison.TwoYearsAgo = &PeriodMetrics{
			Period: PeriodInfo{
				StartDate: startDate.AddDate(-2, 0, 0),
				EndDate:   endDate.AddDate(-2, 0, 0),
			},
			TotalWaterVolume:        twoYearsSummary.TotalWaterVolume,
			TotalEvents:             twoYearsSummary.TotalEvents,
			AverageEfficiency:       twoYearsSummary.AverageEfficiency,
			VolumeChangePercent:     s.calculateChangePercent(currentSummary.TotalWaterVolume, twoYearsSummary.TotalWaterVolume),
			EventsChangePercent:     s.calculateChangePercent(float64(currentSummary.TotalEvents), float64(twoYearsSummary.TotalEvents)),
			EfficiencyChangePercent: s.calculateChangePercent(currentSummary.AverageEfficiency, twoYearsSummary.AverageEfficiency),
		}
	}

	return comparison
}

// calculateSectorBreakdown computes analytics broken down by sector
func (s *analyticsService) calculateSectorBreakdown(farmID uint, startDate, endDate time.Time, aggregation string) []SectorBreakdown {
	// Fetch data for all sectors (no sector filter)
	data, err := s.repo.GetAggregatedData(farmID, nil, startDate, endDate, aggregation)
	if err != nil {
		return []SectorBreakdown{}
	}

	// Group data by sector
	sectorMap := make(map[uint]*SectorBreakdown)

	for _, item := range data {
		d := item.Data
		sectorID := d.IrrigationSectorID
		if breakdown, exists := sectorMap[sectorID]; exists {
			// Update existing sector breakdown
			breakdown.TotalWaterVolume += d.WaterVolume
			breakdown.TotalEvents += item.EventCount // Sum event counts
			breakdown.TotalRealAmount += d.RealAmount
			breakdown.TotalNominalAmount += d.NominalAmount
		} else {
			// Create new sector breakdown
			efficiency := s.calculateEfficiency(d.RealAmount, d.NominalAmount)
			if efficiency == 0 && d.WaterVolume > 0 && d.Duration > 0 {
				nominalVolume := float64(d.Duration) * 1.0
				efficiency = s.calculateEfficiency(d.WaterVolume, nominalVolume)
			}

			sectorMap[sectorID] = &SectorBreakdown{
				SectorID:           sectorID,
				TotalWaterVolume:   d.WaterVolume,
				TotalEvents:        item.EventCount, // Use event count from aggregation
				AverageEfficiency:  efficiency,
				TotalRealAmount:    d.RealAmount,
				TotalNominalAmount: d.NominalAmount,
			}
		}
	}

	// Calculate average efficiency for each sector
	breakdowns := make([]SectorBreakdown, 0, len(sectorMap))
	for _, breakdown := range sectorMap {
		// Recalculate average efficiency based on total real/nominal amounts
		if breakdown.TotalNominalAmount > 0 {
			breakdown.AverageEfficiency = s.calculateEfficiency(breakdown.TotalRealAmount, breakdown.TotalNominalAmount)
		}

		// Round values
		breakdown.TotalWaterVolume = math.Round(breakdown.TotalWaterVolume*100) / 100
		breakdown.TotalRealAmount = math.Round(breakdown.TotalRealAmount*100) / 100
		breakdown.TotalNominalAmount = math.Round(breakdown.TotalNominalAmount*100) / 100
		breakdown.AverageEfficiency = math.Round(breakdown.AverageEfficiency*10000) / 10000

		breakdowns = append(breakdowns, *breakdown)
	}

	return breakdowns
}

// calculateYearOverYear computes YoY comparisons (legacy format)
func (s *analyticsService) calculateYearOverYear(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string, currentSummary AnalyticsSummary) YearOverYearComparison {
	yoy := YearOverYearComparison{}

	// Fetch data for -1 year
	oneYearData, err := s.repo.GetYearOverYearData(farmID, sectorID, startDate, endDate, aggregation, 1)
	if err == nil && len(oneYearData) > 0 {
		oneYearSummary := s.calculateSummary(oneYearData)
		changePercent := s.calculateChangePercent(currentSummary.TotalWaterVolume, oneYearSummary.TotalWaterVolume)

		yoy.OneYearAgo = &YearComparison{
			Period: PeriodInfo{
				StartDate: startDate.AddDate(-1, 0, 0),
				EndDate:   endDate.AddDate(-1, 0, 0),
			},
			TotalWaterVolume:  oneYearSummary.TotalWaterVolume,
			TotalDuration:     oneYearSummary.TotalDuration,
			AverageEfficiency: oneYearSummary.AverageEfficiency,
			TotalEvents:       oneYearSummary.TotalEvents,
			ChangePercent:     changePercent,
		}
	}

	// Fetch data for -2 years
	twoYearsData, err := s.repo.GetYearOverYearData(farmID, sectorID, startDate, endDate, aggregation, 2)
	if err == nil && len(twoYearsData) > 0 {
		twoYearsSummary := s.calculateSummary(twoYearsData)
		changePercent := s.calculateChangePercent(currentSummary.TotalWaterVolume, twoYearsSummary.TotalWaterVolume)

		yoy.TwoYearsAgo = &YearComparison{
			Period: PeriodInfo{
				StartDate: startDate.AddDate(-2, 0, 0),
				EndDate:   endDate.AddDate(-2, 0, 0),
			},
			TotalWaterVolume:  twoYearsSummary.TotalWaterVolume,
			TotalDuration:     twoYearsSummary.TotalDuration,
			AverageEfficiency: twoYearsSummary.AverageEfficiency,
			TotalEvents:       twoYearsSummary.TotalEvents,
			ChangePercent:     changePercent,
		}
	}

	return yoy
}

// calculateChangePercent calculates percentage change between two values
// Handles division by zero and missing data gracefully
func (s *analyticsService) calculateChangePercent(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			// Both are zero - no change
			return 0.0
		}
		// Previous is 0 but current is not - represents infinite growth
		// Return a large positive change (100%) to indicate significant increase
		return 100.0
	}
	change := ((current - previous) / previous) * 100
	return math.Round(change*100) / 100 // Round to 2 decimal places
}
