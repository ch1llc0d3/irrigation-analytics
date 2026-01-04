package repository

import (
	"time"

	"irrigation-analytics/internal/model"

	"gorm.io/gorm"
)

// AggregatedResult represents the result of an aggregation query
type AggregatedResult struct {
	StartTime          time.Time `gorm:"column:start_time"`
	WaterVolume        float64   `gorm:"column:water_volume"`
	Duration           int       `gorm:"column:duration"`
	FarmID             uint      `gorm:"column:farm_id"`
	IrrigationSectorID uint      `gorm:"column:irrigation_sector_id"`
	EventCount         int       `gorm:"column:event_count"`
	NominalAmount      float64   `gorm:"column:nominal_amount"`
	RealAmount         float64   `gorm:"column:real_amount"`
}

// AggregatedDataWithCount wraps IrrigationData with event count
type AggregatedDataWithCount struct {
	Data       model.IrrigationData
	EventCount int
}

// IrrigationRepository defines the interface for irrigation data operations
type IrrigationRepository interface {
	FarmExists(farmID uint) (bool, error)
	GetAggregatedData(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string) ([]AggregatedDataWithCount, error)
	GetYearOverYearData(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string, yearsBack int) ([]AggregatedDataWithCount, error)
}

// irrigationRepository implements IrrigationRepository
type irrigationRepository struct {
	db *gorm.DB
}

// NewIrrigationRepository creates a new irrigation repository
func NewIrrigationRepository(db *gorm.DB) IrrigationRepository {
	return &irrigationRepository{db: db}
}

// FarmExists checks if a farm with the given ID exists
func (r *irrigationRepository) FarmExists(farmID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Farm{}).Where("id = ?", farmID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAggregatedData fetches irrigation data with efficient SQL grouping
func (r *irrigationRepository) GetAggregatedData(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string) ([]AggregatedDataWithCount, error) {
	var results []AggregatedResult
	var modelResults []AggregatedDataWithCount

	// Build base query
	baseQuery := "farm_id = ? AND start_time >= ? AND start_time < ?"
	args := []interface{}{farmID, startDate, endDate}

	if sectorID != nil {
		baseQuery += " AND irrigation_sector_id = ?"
		args = append(args, *sectorID)
	}

	// Build aggregation query based on level
	var sqlQuery string
	switch aggregation {
	case "daily":
		sqlQuery = `
			SELECT 
				DATE(start_time)::timestamp as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE(start_time), farm_id, irrigation_sector_id
			ORDER BY DATE(start_time) ASC`

	case "weekly":
		sqlQuery = `
			SELECT 
				DATE_TRUNC('week', start_time) as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE_TRUNC('week', start_time), farm_id, irrigation_sector_id
			ORDER BY DATE_TRUNC('week', start_time) ASC`

	case "monthly":
		sqlQuery = `
			SELECT 
				DATE_TRUNC('month', start_time) as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE_TRUNC('month', start_time), farm_id, irrigation_sector_id
			ORDER BY DATE_TRUNC('month', start_time) ASC`

	default:
		// Default to daily
		sqlQuery = `
			SELECT 
				DATE(start_time)::timestamp as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE(start_time), farm_id, irrigation_sector_id
			ORDER BY DATE(start_time) ASC`
	}

	err := r.db.Raw(sqlQuery, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert AggregatedResult to AggregatedDataWithCount
	for _, r := range results {
		modelResults = append(modelResults, AggregatedDataWithCount{
			Data: model.IrrigationData{
				StartTime:          r.StartTime,
				WaterVolume:        r.WaterVolume,
				Duration:           r.Duration,
				FarmID:             r.FarmID,
				IrrigationSectorID: r.IrrigationSectorID,
				NominalAmount:      r.NominalAmount,
				RealAmount:         r.RealAmount,
			},
			EventCount: r.EventCount,
		})
	}

	return modelResults, nil
}

// GetYearOverYearData fetches data from the same period N years back
func (r *irrigationRepository) GetYearOverYearData(farmID uint, sectorID *uint, startDate, endDate time.Time, aggregation string, yearsBack int) ([]AggregatedDataWithCount, error) {
	var results []AggregatedResult
	var modelResults []AggregatedDataWithCount

	// Calculate the date range for the previous year(s)
	yearStart := startDate.AddDate(-yearsBack, 0, 0)
	yearEnd := endDate.AddDate(-yearsBack, 0, 0)

	// Build base query
	baseQuery := "farm_id = ? AND start_time >= ? AND start_time < ?"
	args := []interface{}{farmID, yearStart, yearEnd}

	if sectorID != nil {
		baseQuery += " AND irrigation_sector_id = ?"
		args = append(args, *sectorID)
	}

	// Build aggregation query based on level
	var sqlQuery string
	switch aggregation {
	case "daily":
		sqlQuery = `
			SELECT 
				DATE(start_time)::timestamp as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE(start_time), farm_id, irrigation_sector_id
			ORDER BY DATE(start_time) ASC`

	case "weekly":
		sqlQuery = `
			SELECT 
				DATE_TRUNC('week', start_time) as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE_TRUNC('week', start_time), farm_id, irrigation_sector_id
			ORDER BY DATE_TRUNC('week', start_time) ASC`

	case "monthly":
		sqlQuery = `
			SELECT 
				DATE_TRUNC('month', start_time) as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE_TRUNC('month', start_time), farm_id, irrigation_sector_id
			ORDER BY DATE_TRUNC('month', start_time) ASC`

	default:
		sqlQuery = `
			SELECT 
				DATE(start_time)::timestamp as start_time,
				SUM(water_volume) as water_volume,
				SUM(duration) as duration,
				COUNT(*) as event_count,
				SUM(nominal_amount) as nominal_amount,
				SUM(real_amount) as real_amount,
				farm_id,
				COALESCE(irrigation_sector_id, 0) as irrigation_sector_id
			FROM irrigation_data
			WHERE ` + baseQuery + `
			GROUP BY DATE(start_time), farm_id, irrigation_sector_id
			ORDER BY DATE(start_time) ASC`
	}

	err := r.db.Raw(sqlQuery, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert AggregatedResult to AggregatedDataWithCount
	for _, r := range results {
		modelResults = append(modelResults, AggregatedDataWithCount{
			Data: model.IrrigationData{
				StartTime:          r.StartTime,
				WaterVolume:        r.WaterVolume,
				Duration:           r.Duration,
				FarmID:             r.FarmID,
				IrrigationSectorID: r.IrrigationSectorID,
				NominalAmount:      r.NominalAmount,
				RealAmount:         r.RealAmount,
			},
			EventCount: r.EventCount,
		})
	}

	return modelResults, nil
}
