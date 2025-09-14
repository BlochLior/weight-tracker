package tracker

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var verboseStats bool

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display weight tracking statistics",
	Long: `Display comprehensive statistics about your weight tracking data including:
- Minimum and maximum weights with entry details
- Average weight
- Total number of entries
- Time span from first to last entry
- Weight range (max - min)

Use --verbose to show full entry details instead of just entry IDs.

Examples:
  weight-tracker stats                    # Show basic statistics
  weight-tracker stats --verbose          # Show detailed statistics with full entry info`,
	Run: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVarP(&verboseStats, "verbose", "v", false, "Show full entry details instead of just IDs")
}

func runStats(cmd *cobra.Command, args []string) {
	// Note: args are not used for stats command as all options are handled via flags
	_ = args

	// Create store
	store, err := NewDBStore()
	if err != nil {
		fmt.Printf("Error creating store: %v\n", err)
		return
	}
	defer store.Close()

	// Get all weight entries
	entries, err := store.ListWeights(context.Background(), ListOptions{})
	if err != nil {
		fmt.Printf("Error retrieving weight entries: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No weight entries found.")
		return
	}

	// Calculate statistics
	stats := calculateStatistics(entries)

	// Display statistics
	displayStatistics(stats, verboseStats)
}

type WeightStatistics struct {
	MinWeight      float64
	MinWeightEntry WeightEntry
	MaxWeight      float64
	MaxWeightEntry WeightEntry
	AverageWeight  float64
	TotalEntries   int
	TimeSpan       time.Duration
	WeightRange    float64
	FirstEntry     WeightEntry
	LastEntry      WeightEntry
}

func calculateStatistics(entries []WeightEntry) WeightStatistics {
	if len(entries) == 0 {
		return WeightStatistics{}
	}

	// Sort entries by date to get first and last
	firstEntry := entries[0]
	lastEntry := entries[0]
	minEntry := entries[0]
	maxEntry := entries[0]

	totalWeight := 0.0
	validDates := 0

	for _, entry := range entries {
		// Track min/max weights
		if entry.Weight < minEntry.Weight {
			minEntry = entry
		}
		if entry.Weight > maxEntry.Weight {
			maxEntry = entry
		}

		// Track first/last entries by date (only for valid dates)
		if !entry.Date.IsZero() && entry.Date.Year() > 1 {
			validDates++
			if entry.Date.Before(firstEntry.Date) || firstEntry.Date.IsZero() || firstEntry.Date.Year() <= 1 {
				firstEntry = entry
			}
			if entry.Date.After(lastEntry.Date) || lastEntry.Date.IsZero() || lastEntry.Date.Year() <= 1 {
				lastEntry = entry
			}
		}

		totalWeight += entry.Weight
	}

	// Calculate average
	averageWeight := totalWeight / float64(len(entries))

	// Calculate time span
	var timeSpan time.Duration
	if validDates > 1 {
		timeSpan = lastEntry.Date.Sub(firstEntry.Date)
	}

	// Calculate weight range
	weightRange := maxEntry.Weight - minEntry.Weight

	return WeightStatistics{
		MinWeight:      minEntry.Weight,
		MinWeightEntry: minEntry,
		MaxWeight:      maxEntry.Weight,
		MaxWeightEntry: maxEntry,
		AverageWeight:  averageWeight,
		TotalEntries:   len(entries),
		TimeSpan:       timeSpan,
		WeightRange:    weightRange,
		FirstEntry:     firstEntry,
		LastEntry:      lastEntry,
	}
}

func displayStatistics(stats WeightStatistics, verbose bool) {
	fmt.Println("Weight Tracking Statistics")
	fmt.Println("=========================")

	// Total entries
	fmt.Printf("Total Entries: %d\n", stats.TotalEntries)

	// Average weight
	fmt.Printf("Average Weight: %.2f kg\n", stats.AverageWeight)

	// Weight range
	fmt.Printf("Weight Range: %.2f kg (%.2f - %.2f)\n",
		stats.WeightRange, stats.MinWeight, stats.MaxWeight)

	// Min weight
	fmt.Printf("\nMinimum Weight: %.2f kg", stats.MinWeight)
	if verbose {
		fmt.Printf("\n  Entry: ID=%d, Date=%s, Weight=%.2f kg, Note=%s\n",
			stats.MinWeightEntry.ID,
			stats.MinWeightEntry.Date.Format("2006-01-02"),
			stats.MinWeightEntry.Weight,
			stats.MinWeightEntry.Note)
	} else {
		fmt.Printf(" (Entry ID: %d)\n", stats.MinWeightEntry.ID)
	}

	// Max weight
	fmt.Printf("Maximum Weight: %.2f kg", stats.MaxWeight)
	if verbose {
		fmt.Printf("\n  Entry: ID=%d, Date=%s, Weight=%.2f kg, Note=%s\n",
			stats.MaxWeightEntry.ID,
			stats.MaxWeightEntry.Date.Format("2006-01-02"),
			stats.MaxWeightEntry.Weight,
			stats.MaxWeightEntry.Note)
	} else {
		fmt.Printf(" (Entry ID: %d)\n", stats.MaxWeightEntry.ID)
	}

	// Time span
	if stats.TimeSpan > 0 {
		days := int(stats.TimeSpan.Hours() / 24)
		fmt.Printf("\nTime Span: %d days", days)
		if verbose {
			fmt.Printf("\n  From: %s (Entry ID: %d)\n  To: %s (Entry ID: %d)\n",
				stats.FirstEntry.Date.Format("2006-01-02"), stats.FirstEntry.ID,
				stats.LastEntry.Date.Format("2006-01-02"), stats.LastEntry.ID)
		} else {
			fmt.Printf(" (from Entry ID: %d to Entry ID: %d)\n",
				stats.FirstEntry.ID, stats.LastEntry.ID)
		}
	} else {
		fmt.Println("\nTime Span: Unable to calculate (insufficient valid dates)")
	}
}
