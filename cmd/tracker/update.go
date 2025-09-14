package tracker

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id> [flags]",
	Short: "Update a weight entry by ID",
	Long: `Update an existing weight entry in the database by its ID.

You can find the ID of entries by using the 'list' command.
Only the fields you specify with flags will be updated.

Examples:
  weight-tracker update 1 --weight 75.5                    # Update weight only
  weight-tracker update 2 --weight 80.0 --unit lbs         # Update weight and unit
  weight-tracker update 3 --date 01-01-2025 --note "morning weight"  # Update date and note
  weight-tracker update 4 --weight 70.0 --date 15-06-2025 --unit kg --note "after workout"
`,
	Args: cobra.ExactArgs(1),
	Run:  runUpdate,
}

var updateWeight float64
var updateDate string
var updateUnit string
var updateNote string

func init() {
	updateCmd.Flags().Float64VarP(&updateWeight, "weight", "w", 0, "New weight value")
	updateCmd.Flags().StringVarP(&updateDate, "date", "d", "", "New date (format configurable via DATE_INPUT_FORMAT)")
	updateCmd.Flags().StringVarP(&updateUnit, "unit", "u", "", "New unit (kg, lbs)")
	updateCmd.Flags().StringVarP(&updateNote, "note", "n", "", "New note")
}

// runUpdateInternal contains the core logic and returns errors instead of terminating
func runUpdateInternal(cmd *cobra.Command, args []string) error {
	// Create store instance
	store, err := NewDBStore()
	if err != nil {
		return fmt.Errorf("could not create store: %w", err)
	}
	defer store.Close()

	// Parse ID argument
	idStr := args[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID '%s': must be a number", idStr)
	}

	if id <= 0 {
		return fmt.Errorf("ID must be a positive number, got: %d", id)
	}

	// Get the existing entry
	existingEntry, err := store.GetWeight(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to find weight entry with ID %d: %w", id, err)
	}

	// Show current entry
	fmt.Printf("Current weight entry:\n")
	printWeightEntry(existingEntry)

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

	// Show what will be updated
	fmt.Printf("\nUpdated weight entry:\n")
	printWeightEntry(updatedEntry)

	// Confirm the update
	fmt.Print("Are you sure you want to update this entry? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update the entry
	finalEntry, err := store.UpdateWeight(context.Background(), updatedEntry)
	if err != nil {
		return fmt.Errorf("failed to update weight entry: %w", err)
	}

	fmt.Printf("\nSuccessfully updated weight entry with ID %d.\n", id)
	fmt.Printf("Final entry:\n")
	printWeightEntry(finalEntry)

	return nil
}

// runUpdate is the cobra command wrapper that handles errors appropriately for CLI usage
func runUpdate(cmd *cobra.Command, args []string) {
	if err := runUpdateInternal(cmd, args); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
