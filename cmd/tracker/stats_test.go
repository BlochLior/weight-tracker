package tracker

import (
	"context"
	"testing"
	"time"
)

func TestCalculateStatistics(t *testing.T) {
	// Use a base date for consistency and readability
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		entries  []WeightEntry
		expected WeightStatistics
	}{
		{
			name: "single entry",
			entries: []WeightEntry{
				{ID: 1, Weight: 75.5, Date: baseDate, Unit: "kg", Note: "test"},
			},
			expected: WeightStatistics{
				MinWeight:     75.5,
				MaxWeight:     75.5,
				AverageWeight: 75.5,
				TotalEntries:  1,
				WeightRange:   0.0,
				TimeSpan:      0, // Single entry, no time span
			},
		},
		{
			name: "multiple entries with valid dates",
			entries: []WeightEntry{
				{ID: 1, Weight: 75.5, Date: baseDate, Unit: "kg", Note: "start"},
				{ID: 2, Weight: 76.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "mid"}, // Jan 15
				{ID: 3, Weight: 75.2, Date: baseDate.AddDate(0, 1, 0), Unit: "kg", Note: "end"},  // Feb 1
			},
			expected: WeightStatistics{
				MinWeight:     75.2,
				MaxWeight:     76.0,
				AverageWeight: 75.56666666666666, // (75.5 + 76.0 + 75.2) / 3
				TotalEntries:  3,
				WeightRange:   0.8,                 // 76.0 - 75.2
				TimeSpan:      31 * 24 * time.Hour, // 31 days from Jan 1 to Feb 1
			},
		},
		{
			name: "entries with zero dates",
			entries: []WeightEntry{
				{ID: 1, Weight: 80.0, Date: time.Time{}, Unit: "kg", Note: "zero date"},
				{ID: 2, Weight: 70.0, Date: time.Time{}, Unit: "kg", Note: "zero date"},
			},
			expected: WeightStatistics{
				MinWeight:     70.0,
				MaxWeight:     80.0,
				AverageWeight: 75.0,
				TotalEntries:  2,
				WeightRange:   10.0,
				TimeSpan:      0, // No valid dates
			},
		},
		{
			name: "mixed valid and invalid dates",
			entries: []WeightEntry{
				{ID: 1, Weight: 75.0, Date: time.Time{}, Unit: "kg", Note: "invalid date"},
				{ID: 2, Weight: 76.0, Date: baseDate, Unit: "kg", Note: "valid date"},
				{ID: 3, Weight: 77.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "valid date"}, // Jan 15
			},
			expected: WeightStatistics{
				MinWeight:     75.0,
				MaxWeight:     77.0,
				AverageWeight: 76.0,
				TotalEntries:  3,
				WeightRange:   2.0,
				TimeSpan:      14 * 24 * time.Hour, // 14 days between valid dates
			},
		},
		{
			name:    "empty entries",
			entries: []WeightEntry{},
			expected: WeightStatistics{
				MinWeight:     0,
				MaxWeight:     0,
				AverageWeight: 0,
				TotalEntries:  0,
				WeightRange:   0,
				TimeSpan:      0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStatistics(tt.entries)

			// Check basic statistics
			if result.MinWeight != tt.expected.MinWeight {
				t.Errorf("MinWeight = %v, want %v", result.MinWeight, tt.expected.MinWeight)
			}
			if result.MaxWeight != tt.expected.MaxWeight {
				t.Errorf("MaxWeight = %v, want %v", result.MaxWeight, tt.expected.MaxWeight)
			}
			// Allow small floating point differences for average weight
			avgDiff := result.AverageWeight - tt.expected.AverageWeight
			if avgDiff < 0 {
				avgDiff = -avgDiff
			}
			if avgDiff > 0.0001 {
				t.Errorf("AverageWeight = %v, want %v", result.AverageWeight, tt.expected.AverageWeight)
			}
			if result.TotalEntries != tt.expected.TotalEntries {
				t.Errorf("TotalEntries = %v, want %v", result.TotalEntries, tt.expected.TotalEntries)
			}
			// Allow small floating point differences
			weightRangeDiff := result.WeightRange - tt.expected.WeightRange
			if weightRangeDiff < 0 {
				weightRangeDiff = -weightRangeDiff
			}
			if weightRangeDiff > 0.0001 {
				t.Errorf("WeightRange = %v, want %v", result.WeightRange, tt.expected.WeightRange)
			}

			// Check time span (allow small differences due to floating point precision)
			if tt.expected.TimeSpan > 0 {
				diff := result.TimeSpan - tt.expected.TimeSpan
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Hour {
					t.Errorf("TimeSpan = %v, want %v (diff: %v)", result.TimeSpan, tt.expected.TimeSpan, diff)
				}
			} else {
				if result.TimeSpan != tt.expected.TimeSpan {
					t.Errorf("TimeSpan = %v, want %v", result.TimeSpan, tt.expected.TimeSpan)
				}
			}

			// Check that min/max entries are correctly identified
			if len(tt.entries) > 0 {
				if result.MinWeightEntry.Weight != result.MinWeight {
					t.Errorf("MinWeightEntry.Weight = %v, want %v", result.MinWeightEntry.Weight, result.MinWeight)
				}
				if result.MaxWeightEntry.Weight != result.MaxWeight {
					t.Errorf("MaxWeightEntry.Weight = %v, want %v", result.MaxWeightEntry.Weight, result.MaxWeight)
				}
			}
		})
	}
}

func TestDisplayStatistics(t *testing.T) {
	// This test mainly ensures the function doesn't panic
	// The actual output formatting is harder to test without capturing stdout
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	stats := WeightStatistics{
		MinWeight:      70.0,
		MinWeightEntry: WeightEntry{ID: 1, Weight: 70.0, Date: baseDate, Unit: "kg", Note: "min"},
		MaxWeight:      80.0,
		MaxWeightEntry: WeightEntry{ID: 2, Weight: 80.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "max"}, // Jan 15
		AverageWeight:  75.0,
		TotalEntries:   2,
		TimeSpan:       14 * 24 * time.Hour,
		WeightRange:    10.0,
		FirstEntry:     WeightEntry{ID: 1, Weight: 70.0, Date: baseDate, Unit: "kg", Note: "first"},
		LastEntry:      WeightEntry{ID: 2, Weight: 80.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "last"}, // Jan 15
	}

	// Test non-verbose mode
	displayStatistics(stats, false)

	// Test verbose mode
	displayStatistics(stats, true)

	// Test with zero time span
	stats.TimeSpan = 0
	displayStatistics(stats, false)
}

func TestStatsCommandIntegration(t *testing.T) {
	// This is a basic integration test to ensure the stats command can run
	// without errors when given valid data
	store := NewMockStore()
	ctx := context.Background()
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add test entries
	testEntries := []WeightEntry{
		{Weight: 75.5, Date: baseDate, Unit: "kg", Note: "test1"},
		{Weight: 76.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "test2"}, // Jan 15
		{Weight: 75.2, Date: baseDate.AddDate(0, 1, 0), Unit: "kg", Note: "test3"},  // Feb 1
	}

	for _, entry := range testEntries {
		_, err := store.AddWeight(ctx, entry)
		if err != nil {
			t.Fatal(failedTestEntryAdditionString(err))
		}
	}

	// Get entries and calculate statistics
	entries, err := store.ListWeights(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	stats := calculateStatistics(entries)

	// Verify basic statistics
	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 entries, got %d", stats.TotalEntries)
	}
	if stats.MinWeight != 75.2 {
		t.Errorf("Expected min weight 75.2, got %v", stats.MinWeight)
	}
	if stats.MaxWeight != 76.0 {
		t.Errorf("Expected max weight 76.0, got %v", stats.MaxWeight)
	}
	// Allow small floating point differences
	weightRangeDiff := stats.WeightRange - 0.8
	if weightRangeDiff < 0 {
		weightRangeDiff = -weightRangeDiff
	}
	if weightRangeDiff > 0.0001 {
		t.Errorf("Expected weight range 0.8, got %v", stats.WeightRange)
	}
}

func TestWeightStatisticsStruct(t *testing.T) {
	// Test that the WeightStatistics struct can be created and accessed
	stats := WeightStatistics{
		MinWeight:     70.0,
		MaxWeight:     80.0,
		AverageWeight: 75.0,
		TotalEntries:  2,
		WeightRange:   10.0,
		TimeSpan:      24 * time.Hour,
	}

	// Test field access
	if stats.MinWeight != 70.0 {
		t.Errorf("MinWeight field access failed")
	}
	if stats.MaxWeight != 80.0 {
		t.Errorf("MaxWeight field access failed")
	}
	if stats.AverageWeight != 75.0 {
		t.Errorf("AverageWeight field access failed")
	}
	if stats.TotalEntries != 2 {
		t.Errorf("TotalEntries field access failed")
	}
	if stats.WeightRange != 10.0 {
		t.Errorf("WeightRange field access failed")
	}
	if stats.TimeSpan != 24*time.Hour {
		t.Errorf("TimeSpan field access failed")
	}
}
