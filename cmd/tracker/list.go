package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/BlochLior/weight-tracker/internal/db"
	"github.com/BlochLior/weight-tracker/internal/db/sqlc"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List weight entries",
	Long: `List weight entries from the database with optional filtering and sorting.

Examples:
  weight-tracker list                              # List all entries (default: date desc)
  weight-tracker list --from 2025-01-01           # List entries from date
  weight-tracker list --to 2025-12-31             # List entries until date
  weight-tracker list --from 2025-01-01 --to 2025-12-31  # List entries in date range
  weight-tracker list --limit 10                  # List last 10 entries
  weight-tracker list --sort date                 # Sort by date (ascending)
  weight-tracker list --sort date --desc          # Sort by date (descending)
  weight-tracker list --sort weight --desc        # Sort by weight (descending)
  weight-tracker list --unit kg                   # Filter by unit (future feature)
`,
	Run: runList,
}

func init() {
	listCmd.Flags().StringVarP(&fromDate, "from", "f", "", "Start date for filtering (YYYY-MM-DD)")
	listCmd.Flags().StringVarP(&toDate, "to", "t", "", "End date for filtering (YYYY-MM-DD)")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Maximum number of entries to list (0 = no limit)")
	listCmd.Flags().StringVarP(&sort, "sort", "s", "date", "Field to sort by (date, weight)")
	listCmd.Flags().BoolVarP(&desc, "desc", "d", true, "Sort in descending order")
	listCmd.Flags().StringVarP(&unitFilter, "unit", "u", "", "Filter by unit (kg, lbs) - not implemented yet")
}

var fromDate string
var toDate string
var limit int
var sort string
var desc bool
var unitFilter string

// runListInternal contains the core logic and returns errors instead of terminating
func runListInternal(cmd *cobra.Command, args []string) error {
	db, err := db.OpenDB()
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	// --- 1. Handle Optional Flags ---

	// Get limit (0 means no limit, convert to -1 for SQLite)
	limitValue, _ := cmd.Flags().GetInt("limit")
	if limitValue == 0 {
		limitValue = -1 // SQLite treats -1 as no limit
	}

	// For dates, we use sql.NullString to handle optional values
	var startDate sql.NullString
	if cmd.Flags().Changed("from") {
		startDate.String, _ = cmd.Flags().GetString("from")
		startDate.Valid = true
	}

	var endDate sql.NullString
	if cmd.Flags().Changed("to") {
		endDate.String, _ = cmd.Flags().GetString("to")
		endDate.Valid = true
	}

	// --- 2. Handle Sorting ---
	sortColumn, _ := cmd.Flags().GetString("sort")
	sortDesc, _ := cmd.Flags().GetBool("desc")

	// Validate sort column
	if sortColumn != "date" && sortColumn != "weight" {
		return fmt.Errorf("invalid sort column '%s': must be 'date' or 'weight'", sortColumn)
	}

	// --- 3. Build Params Structure ---
	params := sqlc.ListWeightsDateDescParams{
		StartDate: startDate,
		EndDate:   endDate,
		RowLimit:  int64(limitValue),
	}

	// --- 4. Call the appropriate query based on sort criteria ---
	var weights []sqlc.Weight

	switch {
	case sortColumn == "date" && sortDesc:
		weights, err = queries.ListWeightsDateDesc(context.Background(), params)
	case sortColumn == "date" && !sortDesc:
		dateAscParams := sqlc.ListWeightsDateAscParams{
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
			RowLimit:  params.RowLimit,
		}
		weights, err = queries.ListWeightsDateAsc(context.Background(), dateAscParams)
	case sortColumn == "weight" && sortDesc:
		weightDescParams := sqlc.ListWeightsWeightDescParams{
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
			RowLimit:  params.RowLimit,
		}
		weights, err = queries.ListWeightsWeightDesc(context.Background(), weightDescParams)
	case sortColumn == "weight" && !sortDesc:
		weightAscParams := sqlc.ListWeightsWeightAscParams{
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
			RowLimit:  params.RowLimit,
		}
		weights, err = queries.ListWeightsWeightAsc(context.Background(), weightAscParams)
	default:
		// Default to date desc
		weights, err = queries.ListWeightsDateDesc(context.Background(), params)
	}

	if err != nil {
		return fmt.Errorf("failed to list weights: %w", err)
	}

	// --- 5. Handle empty results ---
	if len(weights) == 0 {
		fmt.Println("No weight entries found.")
		return nil
	}

	// --- 6. Print results ---
	fmt.Printf("Found %d weight entries:\n\n", len(weights))
	for _, w := range weights {
		printWeight(w)
		fmt.Println() // Add spacing between entries
	}

	return nil
}

// runList is the cobra command wrapper that handles errors appropriately for CLI usage
func runList(cmd *cobra.Command, args []string) {
	if err := runListInternal(cmd, args); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
