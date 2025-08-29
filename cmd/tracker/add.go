package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/BlochLior/weight-tracker/internal/db"
	"github.com/BlochLior/weight-tracker/internal/db/sqlc"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add weight float [date]",
	Short: "Add a weight entry to database",
	Long: `Adds a weight entry to the database.
The default date that is associated with the entry is current date.
If optional date is provided, that date will be associated with this entry,
instead of the default`,
	Run: runAdd,
}

var date string
var unit string
var note string

func init() {
	// Persistent flags to be inherited for the 'add' command
	addCmd.Flags().StringVarP(&date, "date", "d", "", "The date of the weight entry (DD-MM-YYYY)")
	addCmd.Flags().StringVarP(&unit, "unit", "u", "", "The unit of measurement (kg, lbs)")
	addCmd.Flags().StringVarP(&note, "note", "n", "", "A note for the weight entry")
}

// runAddInternal contains the core logic and returns errors instead of terminating
func runAddInternal(cmd *cobra.Command, args []string) error {
	db, err := db.OpenDB()
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	if len(args) < 1 {
		return fmt.Errorf("minimal add command needs to have a weight entry")
	}

	weightValue, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("add command needs a float weight to process: %w", err)
	}

	// Check if date flag was provided
	var entryDate sql.NullString
	if cmd.Flags().Changed("date") {
		entryDate.String = date
		entryDate.Valid = true
	}

	// Check if unit flag was provided
	var entryUnit sql.NullString
	if cmd.Flags().Changed("unit") {
		entryUnit.String = unit
		entryUnit.Valid = true
	}

	// Check if the note flag was provided
	var entryNote sql.NullString
	if note != "" {
		entryNote.String = note
		entryNote.Valid = true
	}

	params := sqlc.AddWeightParams{
		Weight: weightValue,
		Date:   entryDate,
		Unit:   entryUnit,
		Note:   entryNote,
	}

	weightEntry, err := queries.AddWeight(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to add weight entry: %w", err)
	}

	// Success - log and print result
	log.Printf("`add` called with args: %v", args)
	printWeight(weightEntry)

	return nil
}

// runAdd is the cobra command wrapper that handles errors appropriately for CLI usage
func runAdd(cmd *cobra.Command, args []string) {
	if err := runAddInternal(cmd, args); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
