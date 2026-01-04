# Database Seeding

## Overview
The seeding functionality is available in `internal/repository/seed.go` and can be triggered via an HTTP endpoint during local development.

## Features
- Generates 2 farms with 3 sectors each (6 sectors total)
- Creates 1,000+ irrigation records spanning 2023-2025
- Ensures data exists for the same date ranges across all three years for YoY comparisons
- Populates `nominal_amount` and `real_amount` fields for efficiency calculations

## Usage

### Via HTTP Endpoint (Development Only)
The seed endpoint is only enabled when:
- `GIN_MODE=debug`, OR
- `ENABLE_SEED_ENDPOINT=true`

**Trigger seeding:**
```bash
curl -X POST http://localhost:8080/dev/seed
```

**Response:**
```json
{
  "message": "database seeded successfully"
}
```

### Via Command Line
You can also use the standalone seed utility:
```bash
go run cmd/seed/main.go
```

## Data Generated
- **Farms**: 2 farms (Green Valley Farm, Sunset Orchard)
- **Sectors**: 3 sectors per farm (6 total)
- **Irrigation Records**: 1,000+ records
- **Date Range**: 2023-01-01 to 2025-12-31
- **Events per Day**: 1-3 events per farm per day
- **Total Records**: Approximately 2,190 - 6,570 records

## Notes
- The seed process **clears existing data** before seeding
- Data is generated with realistic variations:
  - Efficiency factors: 0.7 to 1.3
  - Seasonal variation: 20% more water in summer months
  - Duration: 30 minutes to 4 hours per event
