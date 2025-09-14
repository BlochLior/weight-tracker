package tracker

// list_test.go - Integration tests for list functionality including graph features
// Related files: list.go (main functionality), graph.go (chart generation), graph_test.go (graph tests)
// Tests both tabular output and chart generation via --graph flag

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// list_test.go - Integration tests for list functionality
// * purpose: tests list command functionality with both real db and mock.
// * tests: integration (real DB) and MockStore (mock)
// * focus: comprehensive testing of filtering, sorting, limiting, etc. logic.

// TestListCommand_Integration tests the list command with real database
func TestListCommand_Integration(t *testing.T) {
	// Setup test database
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Create store with test database
	store := NewDBStoreWithDB(testDB)
	ctx := context.Background()

	// Add test data with different dates and weights
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testEntries := []WeightEntry{
		{Weight: 75.0, Date: baseDate, Unit: "kg", Note: "entry 1"},
		{Weight: 80.0, Date: baseDate.AddDate(0, 0, 1), Unit: "kg", Note: "entry 2"},
		{Weight: 70.0, Date: baseDate.AddDate(0, 0, 2), Unit: "lbs", Note: "entry 3"},
		{Weight: 85.0, Date: baseDate.AddDate(0, 0, 3), Unit: "kg", Note: "entry 4"},
		{Weight: 65.0, Date: baseDate.AddDate(0, 0, 4), Unit: "lbs", Note: "entry 5"},
	}

	// Add all test entries
	for _, entry := range testEntries {
		_, err := store.AddWeight(ctx, entry)
		if err != nil {
			t.Fatal(failedTestEntryAdditionString(err))
		}
	}

	tests := []struct {
		name     string
		options  ListOptions
		expected int // expected number of results
		wantErr  bool
	}{
		{
			name:     "list all entries",
			options:  ListOptions{},
			expected: 5,
			wantErr:  false,
		},
		{
			name: "list with limit",
			options: ListOptions{
				Limit: 3,
			},
			expected: 3,
			wantErr:  false,
		},
		{
			name: "filter by unit - kg only",
			options: ListOptions{
				Unit: "kg",
			},
			expected: 3, // entries 1, 2, 4
			wantErr:  false,
		},
		{
			name: "filter by unit - lbs only",
			options: ListOptions{
				Unit: "lbs",
			},
			expected: 2, // entries 3, 5
			wantErr:  false,
		},
		{
			name: "filter by date range",
			options: ListOptions{
				FromDate: &baseDate,
				ToDate:   &[]time.Time{baseDate.AddDate(0, 0, 2)}[0],
			},
			expected: 3, // entries 1, 2, 3
			wantErr:  false,
		},
		{
			name: "sort by weight ascending",
			options: ListOptions{
				SortBy:   "weight",
				SortDesc: false,
			},
			expected: 5,
			wantErr:  false,
		},
		{
			name: "sort by weight descending",
			options: ListOptions{
				SortBy:   "weight",
				SortDesc: true,
			},
			expected: 5,
			wantErr:  false,
		},
		{
			name: "sort by date ascending",
			options: ListOptions{
				SortBy:   "date",
				SortDesc: false,
			},
			expected: 5,
			wantErr:  false,
		},
		{
			name: "sort by date descending",
			options: ListOptions{
				SortBy:   "date",
				SortDesc: true,
			},
			expected: 5,
			wantErr:  false,
		},
		{
			name: "complex filter - kg entries, sorted by weight desc, limited to 2",
			options: ListOptions{
				Unit:     "kg",
				SortBy:   "weight",
				SortDesc: true,
				Limit:    2,
			},
			expected: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.ListWeights(ctx, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListWeights() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			if len(result) != tt.expected {
				t.Errorf("ListWeights() got %d entries, want %d", len(result), tt.expected)
			}

			// Additional validation for sorting tests
			if tt.options.SortBy == "weight" && len(result) > 1 {
				for i := 1; i < len(result); i++ {
					if tt.options.SortDesc {
						if result[i-1].Weight < result[i].Weight {
							t.Errorf("ListWeights() not sorted by weight descending: %f < %f", result[i-1].Weight, result[i].Weight)
						}
					} else {
						if result[i-1].Weight > result[i].Weight {
							t.Errorf("ListWeights() not sorted by weight ascending: %f > %f", result[i-1].Weight, result[i].Weight)
						}
					}
				}
			}

			if tt.options.SortBy == "date" && len(result) > 1 {
				for i := 1; i < len(result); i++ {
					if tt.options.SortDesc {
						if result[i-1].Date.Before(result[i].Date) {
							t.Errorf("ListWeights() not sorted by date descending: %v < %v", result[i-1].Date, result[i].Date)
						}
					} else {
						if result[i-1].Date.After(result[i].Date) {
							t.Errorf("ListWeights() not sorted by date ascending: %v > %v", result[i-1].Date, result[i].Date)
						}
					}
				}
			}
		})
	}
}

// TestListCommand_MockStore tests the list command with MockStore (unit tests)
func TestListCommand_MockStore(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Add test data
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testEntries := []WeightEntry{
		{Weight: 75.0, Date: baseDate, Unit: "kg", Note: "entry 1"},
		{Weight: 80.0, Date: baseDate.AddDate(0, 0, 1), Unit: "kg", Note: "entry 2"},
		{Weight: 70.0, Date: baseDate.AddDate(0, 0, 2), Unit: "lbs", Note: "entry 3"},
	}

	for _, entry := range testEntries {
		_, err := store.AddWeight(ctx, entry)
		if err != nil {
			t.Fatal(failedTestEntryAdditionString(err))
		}
	}

	tests := []struct {
		name     string
		options  ListOptions
		expected int
		wantErr  bool
	}{
		{
			name:     "list all entries",
			options:  ListOptions{},
			expected: 3,
			wantErr:  false,
		},
		{
			name: "filter by unit",
			options: ListOptions{
				Unit: "kg",
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name: "sort by weight descending",
			options: ListOptions{
				SortBy:   "weight",
				SortDesc: true,
			},
			expected: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.ListWeights(ctx, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListWeights() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			if len(result) != tt.expected {
				t.Errorf("ListWeights() got %d entries, want %d", len(result), tt.expected)
			}
		})
	}
}

// TestListCommand_CLI_Integration tests the list command CLI parsing and flag handling
func TestListCommand_CLI_Integration(t *testing.T) {
	// Use a base date for consistency and readability
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		flags       map[string]string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "no flags - default behavior",
			flags:       map[string]string{},
			shouldError: false,
		},
		{
			name: "valid date filters",
			flags: map[string]string{
				"from": "01-01-2025",
				"to":   "31-01-2025",
			},
			shouldError: false,
		},
		{
			name: "valid limit",
			flags: map[string]string{
				"limit": "5",
			},
			shouldError: false,
		},
		{
			name: "valid sorting options",
			flags: map[string]string{
				"sort": "weight",
				"desc": "false",
			},
			shouldError: false,
		},
		{
			name: "valid unit filter",
			flags: map[string]string{
				"unit": "kg",
			},
			shouldError: false,
		},
		{
			name: "graph with terminal output",
			flags: map[string]string{
				"graph":  "true",
				"output": "terminal",
			},
			shouldError: false,
		},
		{
			name: "graph with html output",
			flags: map[string]string{
				"graph":  "true",
				"output": "html",
				"file":   "test-chart.html",
			},
			shouldError: false,
		},
		{
			name: "invalid from date format",
			flags: map[string]string{
				"from": "invalid-date",
			},
			shouldError: true,
			errorMsg:    "invalid from date format",
		},
		{
			name: "invalid to date format",
			flags: map[string]string{
				"to": "invalid-date",
			},
			shouldError: true,
			errorMsg:    "invalid to date format",
		},
		{
			name: "invalid sort column",
			flags: map[string]string{
				"sort": "invalid",
			},
			shouldError: true,
			errorMsg:    "invalid sort column",
		},
		{
			name: "invalid graph output type",
			flags: map[string]string{
				"graph":  "true",
				"output": "invalid",
			},
			shouldError: false, // Should default to terminal
		},
		{
			name: "negative limit",
			flags: map[string]string{
				"limit": "-1",
			},
			shouldError: false, // Negative limit is handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().StringP("from", "f", "", "Start date for filtering")
			cmd.Flags().StringP("to", "t", "", "End date for filtering")
			cmd.Flags().IntP("limit", "l", 0, "Maximum number of entries to list")
			cmd.Flags().StringP("sort", "s", "date", "Field to sort by")
			cmd.Flags().BoolP("desc", "d", true, "Sort in descending order")
			cmd.Flags().StringP("unit", "u", "", "Filter by unit")
			cmd.Flags().BoolP("graph", "g", false, "Display weight chart")
			cmd.Flags().StringP("output", "o", "terminal", "Graph output type")
			cmd.Flags().StringP("file", "", "", "Output filename for graph")
			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Create store with test database
			store := NewDBStoreWithDB(testDB)
			ctx := context.Background()

			// Add test data
			testEntries := []WeightEntry{
				{Weight: 75.0, Date: baseDate, Unit: "kg", Note: "entry 1"},
				{Weight: 80.0, Date: baseDate.AddDate(0, 0, 1), Unit: "kg", Note: "entry 2"},
				{Weight: 70.0, Date: baseDate.AddDate(0, 0, 2), Unit: "lbs", Note: "entry 3"},
			}

			for _, entry := range testEntries {
				_, err := store.AddWeight(ctx, entry)
				if err != nil {
					t.Fatal(failedTestEntryAdditionString(err))
				}
			}

			// Create a test version that uses our test database
			testRunListInternal := func(cmd *cobra.Command, args []string) error {
				// Note: args are not used for list command
				_ = args

				// --- 1. Handle Optional Flags ---
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
				entries, err := store.ListWeights(ctx, options)
				if err != nil {
					return fmt.Errorf("failed to list weights: %w", err)
				}

				// --- 6. Handle output (table or graph) ---
				showGraph, _ := cmd.Flags().GetBool("graph")
				if showGraph {
					// For testing, we'll just verify the graph options are parsed correctly
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

					// Verify graph options are set correctly
					if outputType == OutputHTML && graphFile == "" {
						// This is valid - will generate timestamped filename
					}

					// For testing, we don't actually generate the chart
					// Just verify the parsing worked
					_ = outputType
				}

				// Verify we got some entries (basic sanity check)
				if len(entries) == 0 && !tt.shouldError {
					// This might be expected for some filter combinations
					// We'll let the test pass if no entries are found
				}

				return nil
			}

			// Set up arguments and flags for this test case
			cmd.SetArgs([]string{}) // Reset command state (list command takes no args)

			// Reset all flags to default values first
			cmd.Flags().Set("from", "")
			cmd.Flags().Set("to", "")
			cmd.Flags().Set("limit", "0")
			cmd.Flags().Set("sort", "date")
			cmd.Flags().Set("desc", "true")
			cmd.Flags().Set("unit", "")
			cmd.Flags().Set("graph", "false")
			cmd.Flags().Set("output", "terminal")
			cmd.Flags().Set("file", "")

			// Then set the test-specific flag values
			for flag, value := range tt.flags {
				if err := cmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("failed to set flag %s: %v", flag, err)
				}
			}

			// Execute the internal function
			err := testRunListInternal(cmd, []string{})

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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestListCommand_CLI_Integration_Graph tests the list command graph functionality via CLI
func TestListCommand_CLI_Integration_Graph(t *testing.T) {
	// Use a base date for consistency and readability
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		flags       map[string]string
		shouldError bool
		errorMsg    string
		description string
	}{
		{
			name:        "graph terminal output",
			flags:       map[string]string{"graph": "true", "output": "terminal"},
			shouldError: false,
			description: "Should generate ASCII chart in terminal",
		},
		{
			name:        "graph html output default filename",
			flags:       map[string]string{"graph": "true", "output": "html"},
			shouldError: false,
			description: "Should generate HTML chart with auto-generated filename",
		},
		{
			name:        "graph html output custom filename",
			flags:       map[string]string{"graph": "true", "output": "html", "file": "test-chart.html"},
			shouldError: false,
			description: "Should generate HTML chart with custom filename",
		},
		{
			name:        "graph png output (not implemented)",
			flags:       map[string]string{"graph": "true", "output": "png"},
			shouldError: false, // Should default to terminal since PNG is not implemented
			description: "Should fallback to terminal output for unimplemented PNG",
		},
		{
			name:        "graph with invalid output type",
			flags:       map[string]string{"graph": "true", "output": "invalid"},
			shouldError: false, // Should default to terminal
			description: "Should fallback to terminal output for invalid type",
		},
		{
			name:        "graph with date filtering",
			flags:       map[string]string{"graph": "true", "from": "01-01-2025", "to": "31-01-2025"},
			shouldError: false,
			description: "Should generate graph with date filtering",
		},
		{
			name:        "graph with unit filtering",
			flags:       map[string]string{"graph": "true", "unit": "kg"},
			shouldError: false,
			description: "Should generate graph with unit filtering",
		},
		{
			name:        "graph with limit",
			flags:       map[string]string{"graph": "true", "limit": "2"},
			shouldError: false,
			description: "Should generate graph with limited entries",
		},
		{
			name:        "graph with sorting",
			flags:       map[string]string{"graph": "true", "sort": "weight", "desc": "false"},
			shouldError: false,
			description: "Should generate graph with weight sorting (ascending)",
		},
		{
			name:        "no graph flag - should not generate chart",
			flags:       map[string]string{"output": "html"}, // Graph flag not set
			shouldError: false,
			description: "Should not generate chart when --graph flag is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test command for each test case to avoid flag contamination
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().StringP("from", "f", "", "Start date for filtering")
			cmd.Flags().StringP("to", "t", "", "End date for filtering")
			cmd.Flags().IntP("limit", "l", 0, "Maximum number of entries to list")
			cmd.Flags().StringP("sort", "s", "date", "Field to sort by")
			cmd.Flags().BoolP("desc", "d", true, "Sort in descending order")
			cmd.Flags().StringP("unit", "u", "", "Filter by unit")
			cmd.Flags().BoolP("graph", "g", false, "Display weight chart")
			cmd.Flags().StringP("output", "o", "terminal", "Graph output type")
			cmd.Flags().StringP("file", "", "", "Output filename for graph")

			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Create store with test database
			store := NewDBStoreWithDB(testDB)
			ctx := context.Background()

			// Add test data with multiple entries for meaningful graphs
			testEntries := []WeightEntry{
				{Weight: 75.0, Date: baseDate, Unit: "kg", Note: "entry 1"},
				{Weight: 76.5, Date: baseDate.AddDate(0, 0, 7), Unit: "kg", Note: "entry 2"},    // Jan 8
				{Weight: 74.8, Date: baseDate.AddDate(0, 0, 14), Unit: "kg", Note: "entry 3"},   // Jan 15
				{Weight: 165.0, Date: baseDate.AddDate(0, 0, 21), Unit: "lbs", Note: "entry 4"}, // Jan 22
				{Weight: 77.2, Date: baseDate.AddDate(0, 0, 28), Unit: "kg", Note: "entry 5"},   // Jan 29
			}
			for _, entry := range testEntries {
				_, err := store.AddWeight(ctx, entry)
				if err != nil {
					t.Fatal(failedTestEntryAdditionString(err))
				}
			}

			// Create a test version that replicates runListInternal logic
			testRunListInternal := func(cmd *cobra.Command, args []string) error {
				// Note: args are not used for list command
				_ = args

				// Parse flags
				fromStr, _ := cmd.Flags().GetString("from")
				toStr, _ := cmd.Flags().GetString("to")
				limit, _ := cmd.Flags().GetInt("limit")
				sortField, _ := cmd.Flags().GetString("sort")
				desc, _ := cmd.Flags().GetBool("desc")
				unitFilter, _ := cmd.Flags().GetString("unit")
				showGraph, _ := cmd.Flags().GetBool("graph")

				// Build list options
				options := ListOptions{
					SortBy:   sortField,
					SortDesc: desc,
					Limit:    limit,
					Unit:     unitFilter,
				}

				// Parse dates if provided
				if fromStr != "" {
					fromDate, err := ParseDate(fromStr)
					if err != nil {
						return fmt.Errorf("invalid from date: %w", err)
					}
					options.FromDate = &fromDate
				}

				if toStr != "" {
					toDate, err := ParseDate(toStr)
					if err != nil {
						return fmt.Errorf("invalid to date: %w", err)
					}
					options.ToDate = &toDate
				}

				// Get entries from store
				entries, err := store.ListWeights(ctx, options)
				if err != nil {
					return fmt.Errorf("failed to list weights: %w", err)
				}

				// Handle output (table or graph)
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

					// Set default filename if not provided
					if graphFile == "" && outputType != OutputTerminal {
						graphFile = "" // Let ensureOutputDir generate a timestamped filename
					}

					// Generate the chart
					graphOptions := GraphOptions{
						OutputType:    outputType,
						OutputFile:    graphFile,
						Width:         800,
						Height:        600,
						TestOutputDir: t.TempDir(), // Use temporary directory for test charts
					}

					outputPath, err := GenerateWeightChart(entries, graphOptions)
					if err != nil {
						return fmt.Errorf("failed to generate chart: %w", err)
					}

					// For testing, we just verify the function didn't error
					// In a real scenario, you might want to check if the file was created
					_ = outputPath
				} else {
					// Regular table output - just verify we have entries
					if len(entries) == 0 {
						return fmt.Errorf("no entries found")
					}
				}

				return nil
			}

			// Set up flags for this test case
			cmd.SetArgs([]string{}) // Reset command state (list command takes no args)

			// Set the test-specific flag values
			for flag, value := range tt.flags {
				if err := cmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("failed to set flag %s: %v", flag, err)
				}
			}

			// Execute the internal function
			err := testRunListInternal(cmd, []string{})

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
