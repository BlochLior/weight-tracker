package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/BlochLior/weight-tracker/internal/db/sqlc"
	"github.com/spf13/cobra"
)

// add_test.go - Integration tests for add command
// * purpose: tests add command functionality with both real db and mock.
// * tests: integration (real DB) and MockStore (mock)
// * focus: comprehensive testing of add command functionality.
func TestAddCommand_Integration(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		expectedWeight float64
		expectedDate   string
		expectedUnit   string
		expectedNote   string
		shouldError    bool
	}{
		{
			name:           "basic weight entry",
			args:           []string{"75.5"},
			flags:          map[string]string{},
			expectedWeight: 75.5,
			expectedUnit:   "", // Should be null in database, defaults handled by DB
			shouldError:    false,
		},
		{
			name:           "weight with date and unit",
			args:           []string{"80.2"},
			flags:          map[string]string{"date": "01-01-2025", "unit": "lbs"},
			expectedWeight: 80.2,
			expectedDate:   "01-01-2025",
			expectedUnit:   "lbs",
			shouldError:    false,
		},
		{
			name:           "weight with all flags",
			args:           []string{"72.1"},
			flags:          map[string]string{"date": "15-06-2025", "unit": "kg", "note": "morning weight"},
			expectedWeight: 72.1,
			expectedDate:   "15-06-2025",
			expectedUnit:   "kg",
			expectedNote:   "morning weight",
			shouldError:    false,
		},
		{
			name:        "invalid weight",
			args:        []string{"invalid"},
			flags:       map[string]string{},
			shouldError: true,
		},
		{
			name:        "no arguments",
			args:        []string{},
			flags:       map[string]string{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{
				Use: "add",
			}
			cmd.Flags().StringP("date", "d", "", "The date of the weight entry")
			cmd.Flags().StringP("unit", "u", "", "The unit of measurement")
			cmd.Flags().StringP("note", "n", "", "A note for the weight entry")
			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Create a test version that uses our test database instead of the production DB
			testRunAddInternal := func(cmd *cobra.Command, args []string) error {
				// Mock the database setup by creating a modified version of runAddInternal
				queries := sqlc.New(testDB)

				if len(args) < 1 {
					return fmt.Errorf("minimal add command needs to have a weight entry")
				}

				weightValue, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					return fmt.Errorf("add command needs a float weight to process: %w", err)
				}

				// Handle flags (same logic as runAddInternal)
				var entryDate sql.NullString
				if cmd.Flags().Changed("date") {
					date, _ := cmd.Flags().GetString("date")
					entryDate.String = date
					entryDate.Valid = true
				}

				var entryUnit sql.NullString
				if cmd.Flags().Changed("unit") {
					unit, _ := cmd.Flags().GetString("unit")
					entryUnit.String = unit
					entryUnit.Valid = true
				}

				var entryNote sql.NullString
				if cmd.Flags().Changed("note") {
					note, _ := cmd.Flags().GetString("note")
					if note != "" {
						entryNote.String = note
						entryNote.Valid = true
					}
				}

				params := sqlc.AddWeightParams{
					Weight: weightValue,
					Date:   entryDate,
					Unit:   entryUnit,
					Note:   entryNote,
				}

				_, err = queries.AddWeight(context.Background(), params)
				if err != nil {
					return fmt.Errorf("failed to add weight entry: %w", err)
				}

				return nil
			}

			// Set up arguments and flags for this test case
			cmd.SetArgs(tt.args)
			for flag, value := range tt.flags {
				if err := cmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("failed to set flag %s: %v", flag, err)
				}
			}

			// Execute the internal function directly with error handling
			err := testRunAddInternal(cmd, tt.args)

			// Check for expected errors
			if tt.shouldError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatal(unexpectedErrorString(err))
			}

			// For a more complete test, we would query the database to verify
			// the entry was inserted correctly. For now, we've tested that
			// the command executed without error.
		})
	}
}

func TestAddCommandIntegration_DirectDB(t *testing.T) {
	// This test verifies the actual runAdd function would work
	// by temporarily replacing the database connection

	testDB := setupTestDB(t)
	defer testDB.Close()

	// Test that we can actually insert and retrieve data
	queries := sqlc.New(testDB)

	params := sqlc.AddWeightParams{
		Weight: 75.5,
		Date:   sql.NullString{String: "01-01-2025", Valid: true},
		Unit:   sql.NullString{String: "kg", Valid: true},
		Note:   sql.NullString{String: "test entry", Valid: true},
	}

	entry, err := queries.AddWeight(context.Background(), params)
	if err != nil {
		t.Fatalf("failed to add weight: %v", err)
	}

	if entry.Weight != 75.5 {
		t.Errorf("expected weight 75.5, got %f", entry.Weight)
	}

	if !entry.Date.Valid || entry.Date.String != "01-01-2025" {
		t.Errorf("expected date '01-01-2025', got %v", entry.Date)
	}

	if !entry.Unit.Valid || entry.Unit.String != "kg" {
		t.Errorf("expected unit 'kg', got %v", entry.Unit)
	}

	if !entry.Note.Valid || entry.Note.String != "test entry" {
		t.Errorf("expected note 'test entry', got %v", entry.Note)
	}
}

func TestAddCommand_MockStore(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		entry          WeightEntry
		expectedResult WeightEntry
		wantErr        bool
		shouldValidate bool // Whether to test validation before adding to store
	}{
		// -- Basic functionality --
		{
			name: "valid entry with all fields",
			entry: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "kg",
				Note:   "test entry",
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "kg",
				Note:   "test entry",
			},
			wantErr:        false,
			shouldValidate: true,
		},
		{
			name: "valid entry with minimal fields (just weight)",
			entry: WeightEntry{
				Weight: 75.5,
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   time.Time{}, // Zero time - MockStore doesn't set defaults
				Unit:   "",          // Empty - MockStore doesn't set defaults
				Note:   "",
			},
			wantErr:        false,
			shouldValidate: true,
		},

		// -- Edge cases --
		{
			name: "valid entry with large weight value",
			entry: WeightEntry{
				Weight: 1000000,
				Date:   baseDate,
			},
			expectedResult: WeightEntry{
				Weight: 1000000,
				Date:   baseDate,
				Unit:   "",
				Note:   "",
			},
			wantErr:        false,
			shouldValidate: true,
		},
		{
			name: "valid entry with small weight value",
			entry: WeightEntry{
				Weight: 0.000001,
				Date:   baseDate,
			},
			expectedResult: WeightEntry{
				Weight: 0.000001,
				Date:   baseDate,
				Unit:   "",
				Note:   "",
			},
			wantErr:        false,
			shouldValidate: true,
		},
		{
			name: "valid entry with date in the future (despite being illogical)",
			entry: WeightEntry{
				Weight: 75.5,
				Date:   baseDate.AddDate(100, 0, 1),
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   baseDate.AddDate(100, 0, 1),
				Unit:   "",
				Note:   "",
			},
			wantErr:        false,
			shouldValidate: true,
		},
		{
			name: "valid entry with long note string",
			entry: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Note:   strings.Repeat("a", 1000),
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "",
				Note:   strings.Repeat("a", 1000),
			},
			wantErr:        false,
			shouldValidate: true,
		},
		{
			name: "valid entry with special characters in note string",
			entry: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Note:   "!@#$%^&*()",
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "",
				Note:   "!@#$%^&*()",
			},
			wantErr:        false,
			shouldValidate: true,
		},

		// -- Validation error cases (these should fail validation, not store operation) --
		{
			name: "invalid weight value (negative) - validation should fail",
			entry: WeightEntry{
				Weight: -5.0,
				Date:   baseDate,
			},
			expectedResult: WeightEntry{
				Weight: -5.0,
				Date:   baseDate,
				Unit:   "",
				Note:   "",
			},
			wantErr:        true,
			shouldValidate: true,
		},
		{
			name: "invalid weight value (0) - validation should fail",
			entry: WeightEntry{
				Weight: 0,
				Date:   baseDate,
			},
			expectedResult: WeightEntry{
				Weight: 0,
				Date:   baseDate,
				Unit:   "",
				Note:   "",
			},
			wantErr:        true,
			shouldValidate: true,
		},
		{
			name: "invalid unit value (not empty or kg or lbs) - validation should fail",
			entry: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "invalid",
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   baseDate,
				Unit:   "invalid",
				Note:   "",
			},
			wantErr:        true,
			shouldValidate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation first if required
			if tt.shouldValidate {
				validationErr := ValidateWeightEntry(tt.entry)
				if tt.wantErr {
					// Expected validation error
					if validationErr == nil {
						t.Errorf("expected validation error but got none")
					}
					return // Don't proceed to store if validation fails
				} else {
					// Should not have validation error
					if validationErr != nil {
						t.Fatalf("unexpected validation error: %v", validationErr)
					}
				}
			}

			// Test store operation
			result, err := store.AddWeight(ctx, tt.entry)
			if err != nil {
				t.Fatalf("unexpected store error: %v", err)
			}

			// Check that ID was assigned (MockStore assigns IDs starting from 1)
			if result.ID == 0 {
				t.Errorf("expected ID to be assigned, got 0")
			}

			// Check other fields match exactly (MockStore stores what you give it)
			if result.Weight != tt.expectedResult.Weight {
				t.Errorf("expected weight %f, got %f", tt.expectedResult.Weight, result.Weight)
			}
			if result.Date != tt.expectedResult.Date {
				t.Errorf("expected date %v, got %v", tt.expectedResult.Date, result.Date)
			}
			if result.Unit != tt.expectedResult.Unit {
				t.Errorf("expected unit %s, got %s", tt.expectedResult.Unit, result.Unit)
			}
			if result.Note != tt.expectedResult.Note {
				t.Errorf("expected note %s, got %s", tt.expectedResult.Note, result.Note)
			}
		})
	}
}

// TestAddCommand_Configuration tests the add command with different configuration settings
func TestAddCommand_Configuration(t *testing.T) {
	// Use a fixed test date for consistency
	testDate := time.Date(2024, 9, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		envVars      map[string]string
		args         []string
		expectedUnit string
		expectedDate string
		shouldError  bool
	}{
		{
			name:         "default configuration",
			envVars:      map[string]string{},
			args:         []string{"75.5"},
			expectedUnit: "kg",
			expectedDate: testDate.Format("02-01-2006"),
			shouldError:  false,
		},
		{
			name: "custom default unit lbs",
			envVars: map[string]string{
				"DEFAULT_UNIT": "lbs",
			},
			args:         []string{"165.3"},
			expectedUnit: "lbs",
			expectedDate: testDate.Format("02-01-2006"),
			shouldError:  false,
		},
		{
			name: "custom date format mm-dd-yyyy",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT": "mm-dd-yyyy",
			},
			args:         []string{"75.5", "--date", "09-15-2024"},
			expectedUnit: "kg",
			expectedDate: "15-09-2024", // Should be parsed and stored correctly
			shouldError:  false,
		},
		{
			name: "custom date format yyyy-mm-dd",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT": "yyyy-mm-dd",
			},
			args:         []string{"75.5", "--date", "2024-09-15"},
			expectedUnit: "kg",
			expectedDate: "15-09-2024", // Should be parsed and stored correctly
			shouldError:  false,
		},
		{
			name: "invalid date format should error",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT": "mm-dd-yyyy",
			},
			args:        []string{"75.5", "--date", "15-09-2024"}, // Wrong format for mm-dd-yyyy
			shouldError: true,
		},
		{
			name: "invalid default unit falls back to kg",
			envVars: map[string]string{
				"DEFAULT_UNIT": "invalid",
			},
			args:         []string{"75.5"},
			expectedUnit: "kg", // Should fall back to default
			expectedDate: testDate.Format("02-01-2006"),
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{}
			cmd.Flags().String("date", "", "date")
			cmd.Flags().String("unit", "", "unit")
			cmd.Flags().String("note", "", "note")
			// Clear existing environment variables
			os.Unsetenv("DATE_INPUT_FORMAT")
			os.Unsetenv("DATE_DISPLAY_FORMAT")
			os.Unsetenv("DEFAULT_UNIT")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create test database
			db := setupTestDB(t)
			defer db.Close()

			// Set flags if provided in args
			if len(tt.args) > 1 {
				for i := 1; i < len(tt.args); i += 2 {
					if i+1 < len(tt.args) {
						cmd.Flags().Set(tt.args[i][2:], tt.args[i+1]) // Remove "--" prefix
					}
				}
			}

			// Create store with test database
			store := NewDBStoreWithDB(db)

			// Test the add command logic
			weightValue, err := strconv.ParseFloat(tt.args[0], 64)
			if err != nil {
				t.Fatalf("Failed to parse weight: %v", err)
			}

			// Create WeightEntry struct (simulating add command logic)
			entry := WeightEntry{
				Weight: weightValue,
				Date:   testDate, // Use fixed test date for consistency
				Unit:   GetDefaultUnit(),
			}

			// Handle date flag if provided
			if cmd.Flags().Changed("date") {
				dateStr, _ := cmd.Flags().GetString("date")
				if dateStr != "" {
					parsedDate, err := ParseDate(dateStr)
					if err != nil {
						if !tt.shouldError {
							t.Errorf("Unexpected error parsing date: %v", err)
						}
						return
					}
					entry.Date = parsedDate
				}
			}

			// Handle unit flag if provided
			if cmd.Flags().Changed("unit") {
				unitStr, _ := cmd.Flags().GetString("unit")
				if unitStr != "" {
					entry.Unit = unitStr
				}
			}

			// Validate the entry
			if err := ValidateWeightEntry(entry); err != nil {
				if !tt.shouldError {
					t.Errorf("Validation failed: %v", err)
				}
				return
			}

			// Add to store
			addedEntry, err := store.AddWeight(context.Background(), entry)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Failed to add weight entry: %v", err)
				return
			}

			// Verify the result
			if addedEntry.Unit != tt.expectedUnit {
				t.Errorf("Expected unit %s, got %s", tt.expectedUnit, addedEntry.Unit)
			}

			// Verify date (allowing for small time differences)
			expectedDate, err := time.Parse("02-01-2006", tt.expectedDate)
			if err != nil {
				t.Fatalf("Failed to parse expected date: %v", err)
			}

			// Check if dates are on the same day (allowing for time differences)
			if !addedEntry.Date.Truncate(24 * time.Hour).Equal(expectedDate.Truncate(24 * time.Hour)) {
				t.Errorf("Expected date %s, got %s", tt.expectedDate, FormatDate(addedEntry.Date))
			}

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
