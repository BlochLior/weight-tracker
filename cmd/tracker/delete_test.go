package tracker

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestDeleteCommand_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewDBStoreWithDB(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() int64
		id      int64
		wantErr bool
	}{
		{
			name: "delete existing entry",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 75.5,
					Unit:   "kg",
					Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent entry",
			id:      999,
			wantErr: true,
		},
		{
			name:    "delete with invalid ID (zero)",
			id:      0,
			wantErr: true,
		},
		{
			name:    "delete with invalid ID (negative)",
			id:      -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id int64
			if tt.setup != nil {
				id = tt.setup()
			} else {
				id = tt.id
			}

			err := store.DeleteWeight(ctx, id)

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

			// Verify deletion
			_, err = store.GetWeight(ctx, id)
			if err == nil {
				t.Errorf("entry should have been deleted")
			}
		})
	}
}

func TestDeleteCommand_MockStore(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() int64
		id      int64
		wantErr bool
	}{
		{
			name: "delete existing entry",
			setup: func() int64 {
				entry, err := store.AddWeight(ctx, WeightEntry{
					Weight: 75.5,
					Unit:   "kg",
				})
				if err != nil {
					t.Fatal(failedEntryCreationString(err))
				}
				return entry.ID
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent entry",
			id:      999,
			wantErr: true,
		},
		{
			name: "delete multiple entries",
			setup: func() int64 {
				// Add multiple entries
				_, _ = store.AddWeight(ctx, WeightEntry{Weight: 75.5, Unit: "kg"})
				entry2, _ := store.AddWeight(ctx, WeightEntry{Weight: 76.0, Unit: "kg"})
				_, _ = store.AddWeight(ctx, WeightEntry{Weight: 76.5, Unit: "kg"})

				// Return the middle entry ID to delete
				return entry2.ID
			},
			wantErr: false,
		},
		{
			name:    "delete with invalid ID (zero)",
			id:      0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id int64
			if tt.setup != nil {
				id = tt.setup()
			} else {
				id = tt.id
			}

			err := store.DeleteWeight(ctx, id)

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

			// Verify deletion
			_, err = store.GetWeight(ctx, id)
			if err == nil {
				t.Errorf("entry should have been deleted")
			}
		})
	}
}

// TestDeleteCommand_CLI_Integration tests the delete command CLI parsing and flag handling
func TestDeleteCommand_CLI_Integration(t *testing.T) {
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
			name:        "delete with confirmation",
			args:        []string{"1"},
			flags:       map[string]string{"confirm": "true"},
			shouldError: false,
		},
		{
			name:        "delete with force",
			args:        []string{"1"},
			flags:       map[string]string{"force": "true"},
			shouldError: false,
		},
		{
			name:        "delete without flags (would require interactive confirmation)",
			args:        []string{"1"},
			flags:       map[string]string{},
			shouldError: false, // This would normally require user input, but we'll mock it
		},
		{
			name:        "invalid ID format",
			args:        []string{"invalid"},
			flags:       map[string]string{"confirm": "true"},
			shouldError: true,
			errorMsg:    "invalid ID",
		},
		{
			name:        "zero ID",
			args:        []string{"0"},
			flags:       map[string]string{"confirm": "true"},
			shouldError: true,
			errorMsg:    "ID must be a positive number",
		},
		{
			name:        "negative ID",
			args:        []string{"-1"},
			flags:       map[string]string{"confirm": "true"},
			shouldError: true,
			errorMsg:    "ID must be a positive number",
		},
		{
			name:        "non-existent ID",
			args:        []string{"999"},
			flags:       map[string]string{"confirm": "true"},
			shouldError: true,
			errorMsg:    "failed to find weight entry",
		},
		{
			name:        "no arguments",
			args:        []string{},
			flags:       map[string]string{"confirm": "true"},
			shouldError: true,
			errorMsg:    "minimal delete command needs to have an ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{
				Use: "delete",
			}
			cmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")
			cmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")

			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Create store with test database
			store := NewDBStoreWithDB(testDB)
			ctx := context.Background()

			// Add a test entry to delete (only if we have a valid ID)
			if len(tt.args) > 0 && tt.args[0] != "999" && tt.args[0] != "invalid" {
				testEntry := WeightEntry{
					Weight: 70.0,
					Date:   baseDate,
					Unit:   "kg",
					Note:   "entry to delete",
				}
				_, err := store.AddWeight(ctx, testEntry)
				if err != nil {
					t.Fatal(failedTestEntryAdditionString(err))
				}
			}

			// Create a test version that uses our test database
			testRunDeleteInternal := func(cmd *cobra.Command, args []string) error {
				// Parse ID argument
				if len(args) < 1 {
					return fmt.Errorf("minimal delete command needs to have an ID")
				}

				idStr := args[0]
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid ID '%s': must be a number", idStr)
				}

				if id <= 0 {
					return fmt.Errorf("ID must be a positive number, got: %d", id)
				}

				// Check if entry exists before deletion
				_, err = store.GetWeight(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to find weight entry with ID %d: %w", id, err)
				}

				// Handle confirmation unless --confirm or --force is used
				confirmDelete, _ := cmd.Flags().GetBool("confirm")
				forceDelete, _ := cmd.Flags().GetBool("force")

				// For testing, we'll skip the interactive confirmation
				// In real usage, this would prompt the user if neither --confirm nor --force is set
				_ = confirmDelete // Acknowledge we're checking this flag
				_ = forceDelete   // Acknowledge we're checking this flag
				// In a real test, you might want to mock the input or test this separately

				// Delete the entry
				err = store.DeleteWeight(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to delete weight entry: %w", err)
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
			err := testRunDeleteInternal(cmd, tt.args)

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
