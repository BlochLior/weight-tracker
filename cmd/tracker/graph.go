package tracker

// graph.go - Chart generation functionality for weight tracking data
// Related files: list.go (uses graph functionality via --graph flag), graph_test.go (tests)
// Provides ASCII, HTML, and PNG chart generation using go-echarts library
// Note: PNG is not yet implemented.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// GraphOutputType represents the type of graph output
type GraphOutputType string

const (
	OutputTerminal GraphOutputType = "terminal"
	OutputHTML     GraphOutputType = "html"
	OutputPNG      GraphOutputType = "png"
)

// GraphOptions represents options for graph generation
type GraphOptions struct {
	OutputType GraphOutputType
	OutputFile string
	Width      int
	Height     int
	Title      string
}

// ensureOutputDir creates the output directory if it doesn't exist
func ensureOutputDir(filename string) (string, error) {
	outputDir := "charts"

	// If filename is provided, use it; otherwise generate a default
	if filename == "" {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename = fmt.Sprintf("weight-chart_%s.html", timestamp)
	}

	// Ensure filename has proper extension
	if filepath.Ext(filename) == "" {
		filename += ".html"
	}

	// Create the full path including any nested directories
	fullPath := filepath.Join(outputDir, filename)
	dir := filepath.Dir(fullPath)

	// Create output directory (including nested directories) if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return fullPath, nil
}

// GenerateWeightChart generates a chart from weight entries
func GenerateWeightChart(entries []WeightEntry, options GraphOptions) (string, error) {
	if len(entries) == 0 {
		return "", fmt.Errorf("no weight entries to display")
	}

	// Sort entries by date for proper chronological display
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})

	switch options.OutputType {
	case OutputTerminal:
		err := generateASCIIChart(entries, options)
		return "", err
	case OutputHTML:
		return generateHTMLChart(entries, options)
	case OutputPNG:
		return generatePNGChart(entries, options)
	default:
		return "", fmt.Errorf("unsupported output type: %s", options.OutputType)
	}
}

// generateASCIIChart creates a simple ASCII art chart for terminal display
func generateASCIIChart(entries []WeightEntry, options GraphOptions) error {
	if len(entries) == 0 {
		return fmt.Errorf("no entries to display")
	}

	// Find min and max weights for scaling
	minWeight, maxWeight := entries[0].Weight, entries[0].Weight
	for _, entry := range entries {
		if entry.Weight < minWeight {
			minWeight = entry.Weight
		}
		if entry.Weight > maxWeight {
			maxWeight = entry.Weight
		}
	}

	// Add some padding
	weightRange := maxWeight - minWeight
	if weightRange == 0 {
		weightRange = 1 // Avoid division by zero
	}
	padding := weightRange * 0.1
	minWeight -= padding
	maxWeight += padding
	weightRange = maxWeight - minWeight

	fmt.Printf("\n%s\n", options.Title)
	fmt.Printf("Weight Chart (%d entries)\n", len(entries))
	fmt.Printf("Range: %.1f - %.1f kg\n\n", minWeight, maxWeight)

	// Simple line chart with dots
	chartHeight := 10
	chartWidth := len(entries)
	if chartWidth > 20 {
		chartWidth = 20 // Limit width for readability
	}

	// Create simple chart
	chart := make([][]string, chartHeight)
	for i := range chart {
		chart[i] = make([]string, chartWidth)
		for j := range chart[i] {
			chart[i][j] = " "
		}
	}

	// Plot data points
	for i, entry := range entries {
		if i >= chartWidth {
			break
		}

		// Calculate position
		x := i
		y := int((maxWeight - entry.Weight) / weightRange * float64(chartHeight-1))

		if y >= 0 && y < chartHeight {
			chart[y][x] = "●"
		}
	}

	// Print chart with Y-axis labels
	for i := 0; i < chartHeight; i++ {
		weight := maxWeight - (float64(i)/float64(chartHeight-1))*weightRange
		fmt.Printf("%6.1f |", weight)

		for j := 0; j < chartWidth; j++ {
			fmt.Printf("%s", chart[i][j])
		}
		fmt.Println()
	}

	// Print X-axis
	fmt.Printf("       +")
	for i := 0; i < chartWidth; i++ {
		fmt.Printf("-")
	}
	fmt.Println()

	// Print entry details in a cleaner format
	fmt.Printf("\nWeight Entries:\n")
	for i, entry := range entries {
		fmt.Printf("  %2d. %s: %.1f %s",
			i+1,
			entry.Date.Format("2006-01-02"),
			entry.Weight,
			entry.Unit)
		if entry.Note != "" {
			fmt.Printf(" (%s)", entry.Note)
		}
		fmt.Println()
	}

	return nil
}

// generateHTMLChart creates an HTML chart using go-echarts
func generateHTMLChart(entries []WeightEntry, options GraphOptions) (string, error) {
	if len(entries) == 0 {
		return "", fmt.Errorf("no entries to display")
	}

	// Prepare data for the chart with smart time-based spacing
	var xAxisData []string
	var yAxisData []opts.LineData
	var validEntries []WeightEntry

	// Check if we have proper dates (not zero dates)
	hasProperDates := false
	for _, entry := range entries {
		if !entry.Date.IsZero() && entry.Date.Year() > 1 {
			hasProperDates = true
			break
		}
	}

	if !hasProperDates {
		// All entries have zero/invalid dates - use simple numbering
		for i, entry := range entries {
			xAxisData = append(xAxisData, fmt.Sprintf("Entry %d", i+1))
			yAxisData = append(yAxisData, opts.LineData{Value: entry.Weight})
		}
	} else {
		// We have proper dates - create time-normalized spacing using numeric X-axis
		// Filter entries with proper dates and sort them
		for _, entry := range entries {
			if !entry.Date.IsZero() && entry.Date.Year() > 1 {
				validEntries = append(validEntries, entry)
			}
		}

		// Sort by date
		sort.Slice(validEntries, func(i, j int) bool {
			return validEntries[i].Date.Before(validEntries[j].Date)
		})

		if len(validEntries) == 0 {
			return "", fmt.Errorf("no valid entries with proper dates")
		}

		// Calculate time-based X-axis positions
		startTime := validEntries[0].Date

		// Create X-axis data with time-based positioning
		for _, entry := range validEntries {
			// Calculate position as days from start
			daysFromStart := entry.Date.Sub(startTime).Hours() / 24
			xAxisData = append(xAxisData, fmt.Sprintf("%.1f", daysFromStart))
			yAxisData = append(yAxisData, opts.LineData{Value: entry.Weight})
		}
	}

	// Create line chart with area filling
	line := charts.NewLine()

	// Create informative subtitle
	var subtitle string
	if hasProperDates && len(validEntries) > 0 {
		startDate := validEntries[0].Date.Format("2006-01-02")
		endDate := validEntries[len(validEntries)-1].Date.Format("2006-01-02")
		subtitle = fmt.Sprintf("Period: %s to %s", startDate, endDate)
	} else {
		subtitle = fmt.Sprintf("Total entries: %d", len(entries))
	}

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1600px", // Further increased width to prevent text cutoff
			Height: "800px",  // Increased height for better visibility
		}),
		charts.WithGridOpts(opts.Grid{
			Left:   "15%", // More left padding to prevent title overlap
			Right:  "10%", // More right padding for X-axis label
			Top:    "20%", // More top padding for title space
			Bottom: "15%", // More bottom padding for X-axis label
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    options.Title, // Don't add entry count here, it's already in the title
			Subtitle: subtitle,
			Left:     "center", // Center the title and subtitle to avoid overlap
			Top:      "5%",     // Position title higher to avoid overlap
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: &[]bool{true}[0],
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: &[]bool{true}[0],
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Days from Start",
			Type: "value",
			Min:  "dataMin",
			Max:  "dataMax",
			AxisLabel: &opts.AxisLabel{
				Show: &[]bool{true}[0],
			},
			SplitLine: &opts.SplitLine{
				Show: &[]bool{false}[0], // Remove grid lines for cleaner look
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			// Remove Y-axis label to prevent overlap with title
		}),
	)

	line.SetXAxis(xAxisData).
		AddSeries("Weight", yAxisData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				Smooth:       &[]bool{true}[0],
				ConnectNulls: &[]bool{true}[0], // Connect points across null values
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color: "#5470c6",
			}),
		)

	// Generate HTML with proper output directory
	outputFile, err := ensureOutputDir(options.OutputFile)
	if err != nil {
		return "", err
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	err = line.Render(f)
	if err != nil {
		return "", fmt.Errorf("failed to render chart: %w", err)
	}

	return outputFile, nil
}

// generatePNGChart creates a PNG chart (placeholder - go-echarts doesn't directly support PNG)
func generatePNGChart(entries []WeightEntry, options GraphOptions) (string, error) {
	// Note: go-echarts generates HTML/JS, not direct PNG
	// For PNG output, we'd need to use a headless browser or different library
	// For now, we'll generate HTML and suggest opening in browser
	options.OutputType = OutputHTML
	return generateHTMLChart(entries, options)
}
