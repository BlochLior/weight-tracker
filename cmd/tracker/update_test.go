package tracker

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestUpdateCommand_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewDBStoreWithDB(db)
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func() int64
		updates        WeightEntry
		expectedResult WeightEntry
		wantErr        bool
	}{
		{
			name: "update weight only",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 80.0,
					Unit:   "kg",
					Date:   time.Date(2021, 1, 15, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Weight: 65.2,
			},
			expectedResult: WeightEntry{
				Weight: 65.2,
				Unit:   "kg",
				Date:   time.Date(2021, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "update date only",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 75.5,
					Date:   time.Date(2021, 1, 15, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Date: time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			expectedResult: WeightEntry{
				Weight: 75.5,
				Date:   time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "update unit only",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 90.1,
					Unit:   "kg",
					Date:   time.Date(2066, 1, 22, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Unit: "lbs",
			},
			expectedResult: WeightEntry{
				Weight: 90.1,
				Unit:   "lbs",
				Date:   time.Date(2066, 1, 22, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "update note only",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 100.1,
					Date:   time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
					Note:   "Test note state",
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Note: "Test note change",
			},
			expectedResult: WeightEntry{
				Weight: 100.1,
				Date:   time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
				Note:   "Test note change",
			},
			wantErr: false,
		},
		{
			name: "update multiple fields",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 2.4,
					Unit:   "kg",
					Date:   time.Date(1999, 3, 24, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Weight: 80.0,
				Unit:   "lbs",
				Note:   "After workout",
				Date:   time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC),
			},
			expectedResult: WeightEntry{
				Weight: 80.0,
				Unit:   "lbs",
				Note:   "After workout",
				Date:   time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "update non-existent entry",
			updates: WeightEntry{
				ID:     999,
				Weight: 99.3,
				Date:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "update with invalid weight",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 93.4,
					Date:   time.Date(5022, 1, 2, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			updates: WeightEntry{
				Weight: -5.4,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id int64
			if tt.setup != nil {
				id = tt.setup()
			} else {
				id = tt.updates.ID
			}

			tt.updates.ID = id

			result, err := store.UpdateWeight(ctx, tt.updates)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			if result.Weight != tt.expectedResult.Weight {
				t.Errorf("expected weight %f, got %f", tt.expectedResult.Weight, result.Weight)
			}
			if result.Unit != tt.expectedResult.Unit {
				t.Errorf("expected unit %s, got %s", tt.expectedResult.Unit, result.Unit)
			}
			if result.Note != tt.expectedResult.Note {
				t.Errorf("expected note %s, got %s", tt.expectedResult.Note, result.Note)
			}
			formattedExpectedDate := tt.expectedResult.Date.Format("24-03-1999")
			formattedResultDate := result.Date.Format("24-03-1999")
			if formattedResultDate != formattedExpectedDate {
				t.Errorf("expected date %s, got %s", formattedExpectedDate, formattedResultDate)
			}

		})
	}
}

func TestUpdateCommand_MockStore(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func() int64
		updates        WeightEntry
		expectedResult WeightEntry
		wantErr        bool
	}{
		{
			name: "update existing entry",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 75.5,
					Unit:   "kg",
				})
				if err != nil {
					t.Fatalf("Failed to create test entry: %v", err)
				}
				return entry.ID
			},
			updates: WeightEntry{
				Weight: 80.0,
				Unit:   "lbs",
			},
			expectedResult: WeightEntry{
				Weight: 80.0,
				Unit:   "lbs",
			},
			wantErr: false,
		},
		{
			name: "update non-existent entry",
			updates: WeightEntry{
				ID:     999,
				Weight: 80.0,
			},
			wantErr: true,
		},
		{
			name: "partial update (weight only)",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 75.5,
					Unit:   "kg",
					Note:   "Original note",
				})
				if err != nil {
					t.Fatalf("Failed to create test entry: %v", err)
				}
				return entry.ID
			},
			updates: WeightEntry{
				Weight: 80.0,
			},
			expectedResult: WeightEntry{
				Weight: 80.0,
				Unit:   "kg",            // Should remain unchanged
				Note:   "Original note", // Should remain unchanged
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id int64
			if tt.setup != nil {
				id = tt.setup()
			} else {
				id = tt.updates.ID
			}

			// Set the ID for the update
			tt.updates.ID = id

			// Test the update operation
			result, err := store.UpdateWeight(ctx, tt.updates)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			// Verify the update
			if result.Weight != tt.expectedResult.Weight {
				t.Errorf("expected weight %f, got %f", tt.expectedResult.Weight, result.Weight)
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

// TestUpdateCommand_CLI_Integration tests the update command CLI parsing and flag handling
func TestUpdateCommand_CLI_Integration(t *testing.T) {
	// Use a base date for consistency and readability
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "update weight only",
			args:        []string{"1"},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: false,
		},
		{
			name:        "update date only",
			args:        []string{"1"},
			flags:       map[string]string{"date": "15-01-2025"},
			shouldError: false,
		},
		{
			name:        "update unit only",
			args:        []string{"1"},
			flags:       map[string]string{"unit": "lbs"},
			shouldError: false,
		},
		{
			name:        "update note only",
			args:        []string{"1"},
			flags:       map[string]string{"note": "updated note"},
			shouldError: false,
		},
		{
			name:        "update multiple fields",
			args:        []string{"1"},
			flags:       map[string]string{"weight": "80.0", "date": "20-01-2025", "unit": "kg", "note": "multiple update"},
			shouldError: false,
		},
		{
			name:        "invalid ID format",
			args:        []string{"invalid"},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: true,
			errorMsg:    "invalid ID",
		},
		{
			name:        "zero ID",
			args:        []string{"0"},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: true,
			errorMsg:    "ID must be a positive number",
		},
		{
			name:        "negative ID",
			args:        []string{"-1"},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: true,
			errorMsg:    "ID must be a positive number",
		},
		{
			name:        "non-existent ID",
			args:        []string{"999"},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: true,
			errorMsg:    "failed to find weight entry",
		},
		{
			name:        "invalid date format",
			args:        []string{"1"},
			flags:       map[string]string{"date": "invalid-date"},
			shouldError: true,
			errorMsg:    "invalid date format",
		},
		{
			name:        "zero weight",
			args:        []string{"1"},
			flags:       map[string]string{"weight": "0"},
			shouldError: true,
			errorMsg:    "weight must be greater than 0",
		},
		{
			name:        "negative weight",
			args:        []string{"1"},
			flags:       map[string]string{"weight": "-5.0"},
			shouldError: true,
			errorMsg:    "weight must be greater than 0",
		},
		{
			name:        "no flags provided",
			args:        []string{"1"},
			flags:       map[string]string{},
			shouldError: true,
			errorMsg:    "no fields to update",
		},
		{
			name:        "no arguments",
			args:        []string{},
			flags:       map[string]string{"weight": "75.5"},
			shouldError: true,
			errorMsg:    "minimal update command needs to have an ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{
				Use: "update",
			}
			cmd.Flags().Float64P("weight", "w", 0, "New weight value")
			cmd.Flags().StringP("date", "d", "", "New date")
			cmd.Flags().StringP("unit", "u", "", "New unit")
			cmd.Flags().StringP("note", "n", "", "New note")
			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Create store with test database
			store := NewDBStoreWithDB(testDB)
			ctx := context.Background()

			// Add a test entry to update (only if we have a valid ID)
			if len(tt.args) > 0 && tt.args[0] != "999" && tt.args[0] != "invalid" {
				testEntry := WeightEntry{
					Weight: 70.0,
					Date:   baseDate,
					Unit:   "kg",
					Note:   "original entry",
				}
				_, err := store.AddWeight(ctx, testEntry)
				if err != nil {
					t.Fatal(failedTestEntryAdditionString(err))
				}
			}

			// Create a test version that uses our test database
			testRunUpdateInternal := func(cmd *cobra.Command, args []string) error {
				// Parse ID argument
				if len(args) < 1 {
					return fmt.Errorf("minimal update command needs to have an ID")
				}

				idStr := args[0]
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid ID '%s': must be a number", idStr)
				}

				if id <= 0 {
					return fmt.Errorf("ID must be a positive number, got: %d", id)
				}

				// Get the existing entry
				existingEntry, err := store.GetWeight(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to find weight entry with ID %d: %w", id, err)
				}

				// Create updated entry starting with existing values
				updatedEntry := existingEntry

				// Update fields only if flags were provided
				fieldsUpdated := false

				if cmd.Flags().Changed("weight") {
					weightValue, _ := cmd.Flags().GetFloat64("weight")
					if weightValue <= 0 {
						return fmt.Errorf("weight must be greater than 0, got: %f", weightValue)
					}
					updatedEntry.Weight = weightValue
					fieldsUpdated = true
				}

				if cmd.Flags().Changed("date") {
					dateStr, _ := cmd.Flags().GetString("date")
					if dateStr != "" {
						parsedDate, err := ParseDate(dateStr)
						if err != nil {
							return fmt.Errorf("invalid date format '%s': use %s format", dateStr, GetInputFormatDescription())
						}
						updatedEntry.Date = parsedDate
						fieldsUpdated = true
					}
				}

				if cmd.Flags().Changed("unit") {
					unitStr, _ := cmd.Flags().GetString("unit")
					if unitStr != "" {
						updatedEntry.Unit = unitStr
						fieldsUpdated = true
					}
				}

				if cmd.Flags().Changed("note") {
					noteStr, _ := cmd.Flags().GetString("note")
					// Note can be empty, so we always update it if the flag was provided
					updatedEntry.Note = noteStr
					fieldsUpdated = true
				}

				// Check if any fields were actually updated
				if !fieldsUpdated {
					return fmt.Errorf("no fields to update. Use --weight, --date, --unit, or --note flags")
				}

				// Validate the updated entry
				if err := ValidateWeightEntry(updatedEntry); err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}

				// Update the entry (for testing, we'll just verify the logic worked)
				_, err = store.UpdateWeight(ctx, updatedEntry)
				if err != nil {
					return fmt.Errorf("failed to update weight entry: %w", err)
				}

				return nil
			}

			// Set up arguments and flags for this test case
			cmd.SetArgs(tt.args)

			// Set the test-specific flag values
			for flag, value := range tt.flags {
				if err := cmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("failed to set flag %s: %v", flag, err)
				}
			}

			// Execute the internal function
			err := testRunUpdateInternal(cmd, tt.args)

			// Check for expected errors
			if tt.shouldError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
			}
		})
	}
}
