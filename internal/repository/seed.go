package repository

import (
	"fmt"
	"math/rand"
	"time"

	"irrigation-analytics/internal/model"

	"gorm.io/gorm"
)

// SeedRepository handles database seeding operations
type SeedRepository struct {
	db *gorm.DB
}

// NewSeedRepository creates a new seed repository
func NewSeedRepository(db *gorm.DB) *SeedRepository {
	return &SeedRepository{db: db}
}

// SeedDatabase seeds the database with farms, sectors, and irrigation data
// Generates data from 2023 to 2025 to ensure YoY comparisons work
func (s *SeedRepository) SeedDatabase() error {
	// Clear existing data (optional - comment out if you want to keep existing data)
	if err := s.clearExistingData(); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Create farms
	farms, err := s.createFarms()
	if err != nil {
		return fmt.Errorf("failed to create farms: %w", err)
	}

	// Create sectors for each farm
	sectors, err := s.createSectors(farms)
	if err != nil {
		return fmt.Errorf("failed to create sectors: %w", err)
	}

	// Create irrigation data spanning 2023-2025
	totalRecords, err := s.createIrrigationData(farms, sectors)
	if err != nil {
		return fmt.Errorf("failed to create irrigation data: %w", err)
	}

	fmt.Printf("✓ Seeded database successfully:\n")
	fmt.Printf("  - Farms: %d\n", len(farms))
	fmt.Printf("  - Sectors: %d\n", len(sectors))
	fmt.Printf("  - Irrigation records: %d\n", totalRecords)

	return nil
}

// clearExistingData removes existing data
func (s *SeedRepository) clearExistingData() error {
	if err := s.db.Exec("TRUNCATE TABLE irrigation_data CASCADE").Error; err != nil {
		return err
	}
	if err := s.db.Exec("TRUNCATE TABLE irrigation_sectors CASCADE").Error; err != nil {
		return err
	}
	if err := s.db.Exec("TRUNCATE TABLE farms CASCADE").Error; err != nil {
		return err
	}
	return nil
}

// createFarms creates farm entities
func (s *SeedRepository) createFarms() ([]model.Farm, error) {
	farms := []model.Farm{
		{
			Name:        "Green Valley Farm",
			Location:    "Valley County, CA",
			TotalArea:   500.0,
			Description: "A large-scale agricultural operation specializing in row crops",
		},
		{
			Name:        "Sunset Orchard",
			Location:    "Orchard Hills, WA",
			TotalArea:   350.0,
			Description: "Family-owned orchard producing premium fruits",
		},
	}

	if err := s.db.Create(&farms).Error; err != nil {
		return nil, err
	}

	return farms, nil
}

// createSectors creates irrigation sectors for each farm
func (s *SeedRepository) createSectors(farms []model.Farm) ([]model.IrrigationSector, error) {
	sectors := []model.IrrigationSector{}

	for _, farm := range farms {
		for i := 1; i <= 3; i++ {
			sector := model.IrrigationSector{
				FarmID:      farm.ID,
				Name:        fmt.Sprintf("Sector %d", i),
				Area:        farm.TotalArea / 3.0,
				Description: fmt.Sprintf("Irrigation sector %d for %s", i, farm.Name),
			}
			sectors = append(sectors, sector)
		}
	}

	if err := s.db.Create(&sectors).Error; err != nil {
		return nil, err
	}

	return sectors, nil
}

// createIrrigationData creates irrigation records from 2023 to 2025
func (s *SeedRepository) createIrrigationData(farms []model.Farm, sectors []model.IrrigationSector) (int, error) {
	// Define date range: 2023-01-01 to 2025-12-31
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	// Create a map of sectors by farm for easy lookup
	sectorsByFarm := make(map[uint][]model.IrrigationSector)
	for _, sector := range sectors {
		sectorsByFarm[sector.FarmID] = append(sectorsByFarm[sector.FarmID], sector)
	}

	totalRecords := 0
	rand.Seed(time.Now().UnixNano())
	batchSize := 100
	batch := []model.IrrigationData{}

	// Generate records for each day over the 3-year period
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		// For each farm
		for _, farm := range farms {
			// Get sectors for this farm
			farmSectors := sectorsByFarm[farm.ID]
			if len(farmSectors) == 0 {
				continue
			}

			// Generate 1-3 irrigation events per day per sector
			// This ensures we get over 1,000 records
			eventsPerDay := rand.Intn(3) + 1

			for i := 0; i < eventsPerDay; i++ {
				// Pick a random sector
				sector := farmSectors[rand.Intn(len(farmSectors))]

				// Generate random start time during the day (between 6 AM and 8 PM)
				hour := rand.Intn(14) + 6 // 6-19
				minute := rand.Intn(60)
				startTime := time.Date(
					currentDate.Year(),
					currentDate.Month(),
					currentDate.Day(),
					hour,
					minute,
					0,
					0,
					time.UTC,
				)

				// Duration between 30 minutes and 4 hours
				durationMinutes := rand.Intn(210) + 30 // 30-240 minutes
				endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)

				// Calculate nominal and real amounts
				// Nominal amount: expected amount based on duration (1 liter per minute)
				nominalAmount := float64(durationMinutes) * 1.0

				// Efficiency factor: 0.7 to 1.3 (some events more/less efficient)
				efficiencyFactor := 0.7 + rand.Float64()*0.6
				realAmount := nominalAmount * efficiencyFactor

				// Add some seasonal variation (more water in summer months)
				month := int(currentDate.Month())
				if month >= 6 && month <= 8 {
					realAmount *= 1.2 // 20% more in summer
				}

				// Water volume is the same as real amount for consistency
				waterVolume := realAmount

				irrigationData := model.IrrigationData{
					FarmID:             farm.ID,
					IrrigationSectorID: sector.ID,
					StartTime:          startTime,
					EndTime:            endTime,
					WaterVolume:        waterVolume,
					Duration:           durationMinutes,
					NominalAmount:      nominalAmount,
					RealAmount:         realAmount,
				}

				batch = append(batch, irrigationData)
				totalRecords++

				// Insert in batches for better performance
				if len(batch) >= batchSize {
					if err := s.db.Create(&batch).Error; err != nil {
						return 0, fmt.Errorf("failed to create irrigation data batch: %w", err)
					}
					batch = []model.IrrigationData{}
				}
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Insert remaining records
	if len(batch) > 0 {
		if err := s.db.Create(&batch).Error; err != nil {
			return 0, fmt.Errorf("failed to create final irrigation data batch: %w", err)
		}
	}

	return totalRecords, nil
}

