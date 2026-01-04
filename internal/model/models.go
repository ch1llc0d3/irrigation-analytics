package model

import (
	"time"

	"gorm.io/gorm"
)

// Farm represents a farm entity in the irrigation analytics platform
type Farm struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name        string  `gorm:"not null;size:255" json:"name"`
	Location    string  `gorm:"size:255" json:"location"`
	TotalArea   float64 `gorm:"type:decimal(10,2)" json:"total_area"`
	Description string  `gorm:"type:text" json:"description"`

	// Relationships
	IrrigationSectors []IrrigationSector `gorm:"foreignKey:FarmID;constraint:OnDelete:CASCADE" json:"irrigation_sectors,omitempty"`
	IrrigationData    []IrrigationData   `gorm:"foreignKey:FarmID;constraint:OnDelete:CASCADE" json:"irrigation_data,omitempty"`
}

// TableName specifies the table name for Farm
func (Farm) TableName() string {
	return "farms"
}

// IrrigationSector represents an irrigation sector within a farm
type IrrigationSector struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	FarmID      uint    `gorm:"not null;index" json:"farm_id"`
	Name        string  `gorm:"not null;size:255" json:"name"`
	Area        float64 `gorm:"type:decimal(10,2)" json:"area"`
	Description string  `gorm:"type:text" json:"description"`

	// Relationships
	Farm           Farm             `gorm:"foreignKey:FarmID" json:"farm,omitempty"`
	IrrigationData []IrrigationData `gorm:"foreignKey:IrrigationSectorID;constraint:OnDelete:CASCADE" json:"irrigation_data,omitempty"`
}

// TableName specifies the table name for IrrigationSector
func (IrrigationSector) TableName() string {
	return "irrigation_sectors"
}

// IrrigationData represents irrigation event data for analytics
type IrrigationData struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Foreign keys with composite indexes for Year-over-Year analytics optimization
	FarmID            uint      `gorm:"not null;index:idx_farm_start_time,priority:1;index:idx_farm_sector_time,priority:1" json:"farm_id"`
	IrrigationSectorID uint      `gorm:"not null;index:idx_sector_start_time,priority:1;index:idx_farm_sector_time,priority:2;column:irrigation_sector_id" json:"irrigation_sector_id"`
	StartTime         time.Time `gorm:"not null;index:idx_farm_start_time,priority:2;index:idx_sector_start_time,priority:2;index:idx_farm_sector_time,priority:3" json:"start_time"`
	EndTime           time.Time `gorm:"not null" json:"end_time"`
	
	// Irrigation metrics
	WaterVolume   float64 `gorm:"type:decimal(10,2);not null" json:"water_volume"`
	Duration      int     `gorm:"not null" json:"duration"` // Duration in minutes
	NominalAmount float64 `gorm:"type:numeric(10,2)" json:"nominal_amount"`
	RealAmount    float64 `gorm:"type:numeric(10,2)" json:"real_amount"`

	// Relationships
	Farm   Farm           `gorm:"foreignKey:FarmID" json:"farm,omitempty"`
	Sector IrrigationSector `gorm:"foreignKey:IrrigationSectorID" json:"sector,omitempty"`
}

// TableName specifies the table name for IrrigationData
func (IrrigationData) TableName() string {
	return "irrigation_data"
}

// BeforeCreate hook to calculate duration if not set
func (id *IrrigationData) BeforeCreate(tx *gorm.DB) error {
	if id.Duration == 0 && !id.StartTime.IsZero() && !id.EndTime.IsZero() {
		id.Duration = int(id.EndTime.Sub(id.StartTime).Minutes())
	}
	return nil
}

