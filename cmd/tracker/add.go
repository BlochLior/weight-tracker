package tracker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add weight float [date]",
	Short: "Add a weight entry to database",
	Long: `Adds a weight entry to the database.
The default date that is associated with the entry is current date.
If optional date is provided, that date will be associated with this entry,
instead of the default.

Examples:
  weight-tracker add 75.5
  weight-tracker add 75.5 --date 15-09-2024
  weight-tracker add 165.3 --unit lbs --note "After workout"
  weight-tracker add 75.5 --date 15-09-2024 --unit kg --note "Morning weight"`,
	Run: runAdd,
}

var date string
var unit string
var note string

func init() {
	// Persistent flags to be inherited for the 'add' command
	addCmd.Flags().StringVarP(&date, "date", "d", "", "The date of the weight entry (format configurable via DATE_INPUT_FORMAT)")
	addCmd.Flags().StringVarP(&unit, "unit", "u", "", "The unit of measurement (kg, lbs) - default configurable via DEFAULT_UNIT")
	addCmd.Flags().StringVarP(&note, "note", "n", "", "A note for the weight entry")
}

// runAddInternal contains the core logic and returns errors instead of terminating
func runAddInternal(cmd *cobra.Command, args []string) error {
	// Create store instance
	store, err := NewDBStore()
	if err != nil {
		return fmt.Errorf("could not create store: %w", err)
	}
	defer store.Close()

	if len(args) < 1 {
		return fmt.Errorf("minimal add command needs to have a weight entry")
	}

	// Parse weight value
	weightValue, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("add command needs a float weight to process: %w", err)
	}

	// Create WeightEntry struct
	entry := WeightEntry{
		Weight: weightValue,
		Date:   time.Now(),       // Default to current time
		Unit:   GetDefaultUnit(), // Default unit from configuration
	}

	// Handle date flag
	if cmd.Flags().Changed("date") {
		dateStr, _ := cmd.Flags().GetString("date")
		if dateStr != "" {
			// Parse date using configured format
			parsedDate, err := ParseDate(dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format '%s': use %s format", dateStr, GetInputFormatDescription())
			}
			entry.Date = parsedDate
		}
	}

	// Handle unit flag
	if cmd.Flags().Changed("unit") {
		unitStr, _ := cmd.Flags().GetString("unit")
		if unitStr != "" {
			entry.Unit = unitStr
		}
	}

	// Handle note flag
	if cmd.Flags().Changed("note") {
		noteStr, _ := cmd.Flags().GetString("note")
		if noteStr != "" {
			entry.Note = noteStr
		}
	}

	// Validate the entry
	if err := ValidateWeightEntry(entry); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Add to store
	addedEntry, err := store.AddWeight(context.Background(), entry)
	if err != nil {
		return fmt.Errorf("failed to add weight entry: %w", err)
	}

	// Success - log and print result
	log.Printf("`add` called with args: %v", args)
	printWeightEntry(addedEntry)

	return nil
}

// runAdd is the cobra command wrapper that handles errors appropriately for CLI usage
func runAdd(cmd *cobra.Command, args []string) {
	if err := runAddInternal(cmd, args); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
