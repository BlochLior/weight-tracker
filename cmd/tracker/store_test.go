package tracker

import (
	"testing"
	"time"
)

// store_test.go - Business logic validation tests
// * purpose: test validation logic and business rules
// * focus: Unit testing of validation functions and business logic

// TestValidateWeightEntry tests the ValidateWeightEntry function
func TestValidateWeightEntry(t *testing.T) {
	tests := []struct {
		name    string
		entry   WeightEntry
		wantErr bool
	}{
		{
			name: "valid entry with kg",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "kg",
			},
			wantErr: false,
		},
		{
			name: "valid entry with lbs",
			entry: WeightEntry{
				Weight: 165.0,
				Unit:   "lbs",
			},
			wantErr: false,
		},
		{
			name: "valid entry with empty unit",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "",
			},
			wantErr: false,
		},
		{
			name: "invalid weight - zero",
			entry: WeightEntry{
				Weight: 0,
				Unit:   "kg",
			},
			wantErr: true,
		},
		{
			name: "invalid weight - negative",
			entry: WeightEntry{
				Weight: -5.0,
				Unit:   "kg",
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid unit - mixed case",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "KG",
			},
			wantErr: true,
		},
		{
			name: "invalid unit - partial match",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "kilograms",
			},
			wantErr: true,
		},
		{
			name: "valid entry with date",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "kg",
				Date:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid entry with note",
			entry: WeightEntry{
				Weight: 75.5,
				Unit:   "kg",
				Note:   "morning weight",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWeightEntry(tt.entry)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateWeightEntry() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
			}
		})
	}
}

// TestWeightEntryStruct tests the WeightEntry struct behavior
func TestWeightEntryStruct(t *testing.T) {
	// Test that WeightEntry can be created with all fields
	entry := WeightEntry{
		ID:     1,
		Weight: 75.5,
		Date:   time.Now(),
		Unit:   "kg",
		Note:   "test entry",
		UserID: "user123",
	}

	if entry.ID != 1 {
		t.Errorf("Expected ID 1, got %d", entry.ID)
	}
	if entry.Weight != 75.5 {
		t.Errorf("Expected weight 75.5, got %f", entry.Weight)
	}
	if entry.Unit != "kg" {
		t.Errorf("Expected unit 'kg', got '%s'", entry.Unit)
	}
	if entry.Note != "test entry" {
		t.Errorf("Expected note 'test entry', got '%s'", entry.Note)
	}
	if entry.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", entry.UserID)
	}
	if entry.Date.IsZero() {
		t.Errorf("Expected Date to be set, got zero time")
	}
}

// TestListOptionsStruct tests the ListOptions struct behavior
func TestListOptionsStruct(t *testing.T) {
	now := time.Now()

	// Test that ListOptions can be created with all fields
	options := ListOptions{
		FromDate: &now,
		ToDate:   &now,
		Limit:    10,
		SortBy:   "date",
		SortDesc: true,
		Unit:     "kg",
	}

	if options.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", options.Limit)
	}
	if options.SortBy != "date" {
		t.Errorf("Expected sort by 'date', got '%s'", options.SortBy)
	}
	if !options.SortDesc {
		t.Errorf("Expected SortDesc to be true")
	}
	if options.Unit != "kg" {
		t.Errorf("Expected unit 'kg', got '%s'", options.Unit)
	}
	if options.FromDate == nil {
		t.Errorf("Expected FromDate to be set")
	}
	if options.ToDate == nil {
		t.Errorf("Expected ToDate to be set")
	}
}
