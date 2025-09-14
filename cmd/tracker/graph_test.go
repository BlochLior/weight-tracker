package tracker

// graph_test.go - Unit tests for chart generation functionality
// Related files: graph.go (main functionality), list.go (uses graph features), list_test.go (integration tests)
// Tests ASCII, HTML, and PNG chart generation, file output, and error handling

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateWeightChart(t *testing.T) {
	// Use a base date for consistency and readability
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create test entries
	testEntries := []WeightEntry{
		{
			ID:     1,
			Weight: 75.5,
			Date:   baseDate,
			Unit:   "kg",
			Note:   "New Year",
		},
		{
			ID:     2,
			Weight: 76.0,
			Date:   baseDate.AddDate(0, 0, 14), // Jan 15
			Unit:   "kg",
			Note:   "Mid January",
		},
		{
			ID:     3,
			Weight: 75.2,
			Date:   baseDate.AddDate(0, 1, 0), // Feb 1
			Unit:   "kg",
			Note:   "February",
		},
	}

	tests := []struct {
		name    string
		entries []WeightEntry
		options GraphOptions
		wantErr bool
	}{
		{
			name:    "generate ASCII chart",
			entries: testEntries,
			options: GraphOptions{
				OutputType: OutputTerminal,
				Title:      "Test Chart",
			},
			wantErr: false,
		},
		{
			name:    "generate HTML chart",
			entries: testEntries,
			options: GraphOptions{
				OutputType: OutputHTML,
				Title:      "Test Chart",
				OutputFile: "test-chart.html",
			},
			wantErr: false,
		},
		{
			name:    "empty entries",
			entries: []WeightEntry{},
			options: GraphOptions{
				OutputType: OutputTerminal,
				Title:      "Empty Chart",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath, err := GenerateWeightChart(tt.entries, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateWeightChart() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			// For HTML output, verify file was created
			if tt.options.OutputType == OutputHTML {
				if outputPath == "" {
					t.Errorf("GenerateWeightChart() expected output path for HTML chart")
					return
				}

				// Check if file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("GenerateWeightChart() HTML file was not created: %s", outputPath)
				}

				// Clean up test file
				defer os.Remove(outputPath)
			}
		})
	}
}

func TestGenerateASCIIChart(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testEntries := []WeightEntry{
		{ID: 1, Weight: 75.5, Date: baseDate, Unit: "kg"},
		{ID: 2, Weight: 76.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg"}, // Jan 15
		{ID: 3, Weight: 75.2, Date: baseDate.AddDate(0, 1, 0), Unit: "kg"},  // Feb 1
	}

	options := GraphOptions{
		Title: "Test ASCII Chart",
	}

	err := generateASCIIChart(testEntries, options)
	if err != nil {
		t.Error(unexpectedErrorString(err))
	}
}

func TestGenerateHTMLChart(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testEntries := []WeightEntry{
		{ID: 1, Weight: 75.5, Date: baseDate, Unit: "kg"},
		{ID: 2, Weight: 76.0, Date: baseDate.AddDate(0, 0, 14), Unit: "kg"}, // Jan 15
		{ID: 3, Weight: 75.2, Date: baseDate.AddDate(0, 1, 0), Unit: "kg"},  // Feb 1
	}

	outputPath, err := generateHTMLChart(testEntries, GraphOptions{
		Title:      "Test HTML Chart",
		OutputFile: "test-html-chart.html",
	})

	if err != nil {
		t.Error(unexpectedErrorString(err))
		return
	}

	if outputPath == "" {
		t.Errorf("generateHTMLChart() expected output path")
		return
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("generateHTMLChart() HTML file was not created: %s", outputPath)
	}

	// Clean up test file
	defer os.Remove(outputPath)
}

func TestEnsureOutputDir(t *testing.T) {
	tests := []struct {
		name       string
		outputFile string
		wantErr    bool
		expectDir  string
	}{
		{
			name:       "custom filename",
			outputFile: "my-chart.html",
			wantErr:    false,
			expectDir:  "charts",
		},
		{
			name:       "empty filename",
			outputFile: "",
			wantErr:    false,
			expectDir:  "charts",
		},
		{
			name:       "nested path",
			outputFile: "subdir/chart.html",
			wantErr:    false,
			expectDir:  "charts/subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath, err := ensureOutputDir(tt.outputFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ensureOutputDir() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Error(unexpectedErrorString(err))
				return
			}

			// Check if the output directory exists
			dir := filepath.Dir(outputPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("ensureOutputDir() directory was not created: %s", dir)
			}

			// Clean up test directory
			defer os.RemoveAll("charts")
		})
	}
}

func TestGraphOptions(t *testing.T) {
	// Test default values
	options := GraphOptions{
		Title: "Test Chart",
	}

	if options.Width == 0 {
		options.Width = 1200
	}
	if options.Height == 0 {
		options.Height = 600
	}

	if options.Width != 1200 {
		t.Errorf("Expected default width 1200, got %d", options.Width)
	}
	if options.Height != 600 {
		t.Errorf("Expected default height 600, got %d", options.Height)
	}
}
