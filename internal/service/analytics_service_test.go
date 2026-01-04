package service

import (
	"testing"
)

// TestCalculateEfficiency tests the calculateEfficiency function
func TestCalculateEfficiency(t *testing.T) {
	// Create a service instance for testing
	service := &analyticsService{}

	tests := []struct {
		name           string
		realAmount     float64
		nominalAmount  float64
		expectedResult float64
		description    string
	}{
		{
			name:           "normal case - perfect efficiency",
			realAmount:     100.0,
			nominalAmount:  100.0,
			expectedResult: 1.0,
			description:    "Real equals nominal, efficiency should be 1.0",
		},
		{
			name:           "normal case - higher efficiency",
			realAmount:     120.0,
			nominalAmount:  100.0,
			expectedResult: 1.2,
			description:    "Real is 20% higher than nominal",
		},
		{
			name:           "normal case - lower efficiency",
			realAmount:     80.0,
			nominalAmount:  100.0,
			expectedResult: 0.8,
			description:    "Real is 20% lower than nominal",
		},
		{
			name:           "normal case - very high efficiency",
			realAmount:     150.0,
			nominalAmount:  100.0,
			expectedResult: 1.5,
			description:    "Real is 50% higher than nominal",
		},
		{
			name:           "normal case - very low efficiency",
			realAmount:     50.0,
			nominalAmount:  100.0,
			expectedResult: 0.5,
			description:    "Real is 50% lower than nominal",
		},
		{
			name:           "division by zero - nominal is zero, real is zero",
			realAmount:     0.0,
			nominalAmount:  0.0,
			expectedResult: 0.0,
			description:    "Both zero, should return 0.0 to avoid division by zero",
		},
		{
			name:           "division by zero - nominal is zero, real is positive",
			realAmount:     100.0,
			nominalAmount:  0.0,
			expectedResult: 0.0,
			description:    "Nominal is zero, should return 0.0 to avoid division by zero",
		},
		{
			name:           "real amount is zero",
			realAmount:     0.0,
			nominalAmount:  100.0,
			expectedResult: 0.0,
			description:    "Real is zero, efficiency should be 0.0",
		},
		{
			name:           "decimal precision - rounds to 4 decimal places",
			realAmount:     100.123456,
			nominalAmount:  100.0,
			expectedResult: 1.0012,
			description:    "Should round to 4 decimal places",
		},
		{
			name:           "small values",
			realAmount:     0.001,
			nominalAmount:  0.01,
			expectedResult: 0.1,
			description:    "Handles very small values correctly",
		},
		{
			name:           "large values",
			realAmount:     1000000.0,
			nominalAmount:  500000.0,
			expectedResult: 2.0,
			description:    "Handles large values correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateEfficiency(tt.realAmount, tt.nominalAmount)
			if result != tt.expectedResult {
				t.Errorf("calculateEfficiency(%f, %f) = %f, expected %f. %s",
					tt.realAmount, tt.nominalAmount, result, tt.expectedResult, tt.description)
			}
		})
	}
}

// TestCalculateChangePercent tests the calculateChangePercent function
func TestCalculateChangePercent(t *testing.T) {
	// Create a service instance for testing
	service := &analyticsService{}

	tests := []struct {
		name           string
		current        float64
		previous       float64
		expectedResult float64
		description    string
	}{
		{
			name:           "normal case - positive change",
			current:        110.0,
			previous:       100.0,
			expectedResult: 10.0,
			description:    "10% increase from 100 to 110",
		},
		{
			name:           "normal case - negative change",
			current:        90.0,
			previous:       100.0,
			expectedResult: -10.0,
			description:    "10% decrease from 100 to 90",
		},
		{
			name:           "normal case - no change",
			current:        100.0,
			previous:       100.0,
			expectedResult: 0.0,
			description:    "No change, should return 0.0",
		},
		{
			name:           "normal case - large increase",
			current:        200.0,
			previous:       100.0,
			expectedResult: 100.0,
			description:    "100% increase (doubled)",
		},
		{
			name:           "normal case - large decrease",
			current:        50.0,
			previous:       100.0,
			expectedResult: -50.0,
			description:    "50% decrease (halved)",
		},
		{
			name:           "division by zero - previous is zero, current is zero",
			current:        0.0,
			previous:       0.0,
			expectedResult: 0.0,
			description:    "Both zero, should return 0.0 (no change)",
		},
		{
			name:           "division by zero - previous is zero, current is positive",
			current:        100.0,
			previous:       0.0,
			expectedResult: 100.0,
			description:    "Previous is zero, current is positive - should return 100.0 (significant increase)",
		},
		{
			name:           "division by zero - previous is zero, current is also zero (edge case)",
			current:        0.0,
			previous:       0.0,
			expectedResult: 0.0,
			description:    "Both zero, should return 0.0",
		},
		{
			name:           "current is zero, previous is positive",
			current:        0.0,
			previous:       100.0,
			expectedResult: -100.0,
			description:    "Current is zero, should return -100.0 (complete decrease)",
		},
		{
			name:           "decimal precision - rounds to 2 decimal places",
			current:        111.111,
			previous:       100.0,
			expectedResult: 11.11,
			description:    "Should round to 2 decimal places",
		},
		{
			name:           "small values",
			current:        0.11,
			previous:       0.10,
			expectedResult: 10.0,
			description:    "Handles very small values correctly",
		},
		{
			name:           "large values",
			current:        2000000.0,
			previous:       1000000.0,
			expectedResult: 100.0,
			description:    "Handles large values correctly",
		},
		{
			name:           "fractional percentage change",
			current:        105.0,
			previous:       100.0,
			expectedResult: 5.0,
			description:    "5% increase",
		},
		{
			name:           "negative previous value (edge case)",
			current:        100.0,
			previous:       -50.0,
			expectedResult: -300.0,
			description:    "Previous is negative, calculates change correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateChangePercent(tt.current, tt.previous)
			if result != tt.expectedResult {
				t.Errorf("calculateChangePercent(%f, %f) = %f, expected %f. %s",
					tt.current, tt.previous, result, tt.expectedResult, tt.description)
			}
		})
	}
}

// TestCalculateChangePercent_DivisionByZero_YoY tests the division by zero case
// specifically for Year-over-Year scenarios where a previous year has 0 volume
func TestCalculateChangePercent_DivisionByZero_YoY(t *testing.T) {
	service := &analyticsService{}

	t.Run("previous year has zero volume, current year has data", func(t *testing.T) {
		// Scenario: Previous year had no irrigation events (0 volume)
		// Current year has 1000 liters
		currentVolume := 1000.0
		previousVolume := 0.0

		result := service.calculateChangePercent(currentVolume, previousVolume)

		// Should return 100.0 to indicate significant increase
		// (can't calculate percentage from zero, so we indicate it's a new occurrence)
		if result != 100.0 {
			t.Errorf("Expected 100.0 for division by zero case (current=%f, previous=%f), got %f",
				currentVolume, previousVolume, result)
		}
	})

	t.Run("previous year has zero volume, current year also zero", func(t *testing.T) {
		// Scenario: Both years have no data
		currentVolume := 0.0
		previousVolume := 0.0

		result := service.calculateChangePercent(currentVolume, previousVolume)

		// Should return 0.0 (no change)
		if result != 0.0 {
			t.Errorf("Expected 0.0 for both zero case (current=%f, previous=%f), got %f",
				currentVolume, previousVolume, result)
		}
	})

	t.Run("previous year has zero events, current year has events", func(t *testing.T) {
		// Scenario: Previous year had 0 events, current year has 50 events
		currentEvents := 50.0
		previousEvents := 0.0

		result := service.calculateChangePercent(currentEvents, previousEvents)

		// Should return 100.0 (significant increase from zero)
		if result != 100.0 {
			t.Errorf("Expected 100.0 for events division by zero case (current=%f, previous=%f), got %f",
				currentEvents, previousEvents, result)
		}
	})

	t.Run("previous year has zero efficiency, current year has efficiency", func(t *testing.T) {
		// Scenario: Previous year had 0 efficiency (no data), current year has 0.85 efficiency
		currentEfficiency := 0.85
		previousEfficiency := 0.0

		result := service.calculateChangePercent(currentEfficiency, previousEfficiency)

		// Should return 100.0 (significant increase from zero)
		if result != 100.0 {
			t.Errorf("Expected 100.0 for efficiency division by zero case (current=%f, previous=%f), got %f",
				currentEfficiency, previousEfficiency, result)
		}
	})

	t.Run("realistic YoY scenario - volume increase from zero", func(t *testing.T) {
		// Scenario: Farm started operations this year
		// Year 1: 0 liters (no data)
		// Year 2: 5000 liters (first year of operations)
		currentYear := 5000.0
		previousYear := 0.0

		result := service.calculateChangePercent(currentYear, previousYear)

		// Should return 100.0 to indicate new operations started
		if result != 100.0 {
			t.Errorf("Expected 100.0 for new operations scenario (current=%f, previous=%f), got %f",
				currentYear, previousYear, result)
		}
	})
}
