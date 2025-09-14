package tracker

// list.go - List command with integrated graph functionality
// Related files: graph.go (chart generation), list_test.go (tests)
// The list command provides both tabular output and chart generation via --graph flag

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List weight entries",
	Long: `List weight entries from the database with optional filtering and sorting.

Examples:
  weight-tracker list                              # List all entries (default: date desc)
  weight-tracker list --from 01-01-2025           # List entries from date
  weight-tracker list --to 31-12-2025             # List entries until date
  weight-tracker list --from 01-01-2025 --to 31-12-2025  # List entries in date range
  weight-tracker list --limit 10                  # List last 10 entries
  weight-tracker list --sort date                 # Sort by date (ascending)
  weight-tracker list --sort date --desc          # Sort by date (descending)
  weight-tracker list --sort weight --desc        # Sort by weight (descending)
  weight-tracker list --unit kg                   # Filter by unit
  weight-tracker list --graph                     # Display ASCII chart in terminal
  weight-tracker list --graph --output html       # Generate HTML chart in charts/ directory
  weight-tracker list --graph --output html --file my-chart.html # Generate HTML chart with custom filename
`,
	Run: runList,
}

func init() {
	listCmd.Flags().StringVarP(&fromDate, "from", "f", "", "Start date for filtering (format configurable via DATE_INPUT_FORMAT)")
	listCmd.Flags().StringVarP(&toDate, "to", "t", "", "End date for filtering (format configurable via DATE_INPUT_FORMAT)")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Maximum number of entries to list (0 = no limit)")
	listCmd.Flags().StringVarP(&sortField, "sort", "s", "date", "Field to sort by (date, weight)")
	listCmd.Flags().BoolVarP(&desc, "desc", "d", true, "Sort in descending order")
	listCmd.Flags().StringVarP(&unitFilter, "unit", "u", "", "Filter by unit (kg, lbs)")
	listCmd.Flags().BoolVarP(&showGraph, "graph", "g", false, "Display weight chart")
	listCmd.Flags().StringVarP(&graphOutput, "output", "o", "terminal", "Graph output type (terminal, html, png)")
	listCmd.Flags().StringVarP(&graphFile, "file", "", "", "Output filename for graph (saved in charts/ directory)")
}

var fromDate string
var toDate string
var limit int
var sortField string
var desc bool
var unitFilter string
var showGraph bool
var graphOutput string
var graphFile string

// runListInternal contains the core logic and returns errors instead of terminating
func runListInternal(cmd *cobra.Command, args []string) error {
	// Note: args are not used for list command as all options are handled via flags
	_ = args
	// Create store instance
	store, err := NewDBStore()
	if err != nil {
		return fmt.Errorf("could not create store: %w", err)
	}
	defer store.Close()

	// --- 1. Handle Optional Flags ---

	// Get limit (0 means no limit)
	limitValue, _ := cmd.Flags().GetInt("limit")

	// Parse date filters
	var fromDate, toDate *time.Time
	if cmd.Flags().Changed("from") {
		dateStr, _ := cmd.Flags().GetString("from")
		if dateStr != "" {
			parsedDate, err := ParseDate(dateStr)
			if err != nil {
				return fmt.Errorf("invalid from date format '%s': use %s format", dateStr, GetInputFormatDescription())
			}
			fromDate = &parsedDate
		}
	}

	if cmd.Flags().Changed("to") {
		dateStr, _ := cmd.Flags().GetString("to")
		if dateStr != "" {
			parsedDate, err := ParseDate(dateStr)
			if err != nil {
				return fmt.Errorf("invalid to date format '%s': use %s format", dateStr, GetInputFormatDescription())
			}
			toDate = &parsedDate
		}
	}

	// --- 2. Handle Sorting ---
	sortColumn, _ := cmd.Flags().GetString("sort")
	sortDesc, _ := cmd.Flags().GetBool("desc")

	// Validate sort column
	if sortColumn != "date" && sortColumn != "weight" {
		return fmt.Errorf("invalid sort column '%s': must be 'date' or 'weight'", sortColumn)
	}

	// --- 3. Handle Unit Filter ---
	unitFilter, _ := cmd.Flags().GetString("unit")

	// --- 4. Build ListOptions ---
	options := ListOptions{
		FromDate: fromDate,
		ToDate:   toDate,
		Limit:    limitValue,
		SortBy:   sortColumn,
		SortDesc: sortDesc,
		Unit:     unitFilter,
	}

	// --- 5. Call the store method ---
	entries, err := store.ListWeights(context.Background(), options)
	if err != nil {
		return fmt.Errorf("failed to list weights: %w", err)
	}

	// --- 6. Handle output (table or graph) ---
	showGraph, _ := cmd.Flags().GetBool("graph")
	if showGraph {
		// Generate graph
		graphOutput, _ := cmd.Flags().GetString("output")
		graphFile, _ := cmd.Flags().GetString("file")

		// Determine output type
		var outputType GraphOutputType
		switch graphOutput {
		case "html":
			outputType = OutputHTML
		case "png":
			outputType = OutputPNG
		default:
			outputType = OutputTerminal
		}

		// Set default filename if not provided (will be handled by ensureOutputDir)
		if graphFile == "" && outputType != OutputTerminal {
			graphFile = "" // Let ensureOutputDir generate a timestamped filename
		}

		// Generate chart title
		title := "Weight Tracking Chart"
		if len(entries) > 0 {
			title = fmt.Sprintf("Weight Tracking Chart (%d entries)", len(entries))
		}

		// Generate the chart
		graphOptions := GraphOptions{
			OutputType: outputType,
			OutputFile: graphFile,
			Width:      800,
			Height:     600,
			Title:      title,
		}

		outputPath, err := GenerateWeightChart(entries, graphOptions)
		if err != nil {
			return fmt.Errorf("failed to generate chart: %w", err)
		}

		// Print success message for file output
		if outputType != OutputTerminal {
			fmt.Printf("Chart generated successfully: %s\n", outputPath)
			if outputType == OutputHTML {
				fmt.Printf("Open %s in your browser to view the chart.\n", outputPath)
			}
		}
	} else {
		// Print table format
		printWeightEntries(entries)
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
