package tracker

import (
	"fmt"
)

// printWeightEntry prints a WeightEntry struct (new function for Store interface)
func printWeightEntry(entry WeightEntry) {
	fmt.Printf("* Weight Entry ID: %d\n", entry.ID)
	fmt.Printf("* Date: %s\n", FormatDate(entry.Date))
	fmt.Printf("* Weight: %.2f %s\n", entry.Weight, entry.Unit)
	if entry.Note != "" {
		fmt.Printf("* Note: %s\n", entry.Note)
	}
	if entry.UserID != "" {
		fmt.Printf("* UserID: %s\n", entry.UserID)
	}
	fmt.Println()
}

// printWeightEntries prints a slice of WeightEntry structs
func printWeightEntries(entries []WeightEntry) {
	if len(entries) == 0 {
		fmt.Println("No weight entries found.")
		return
	}

	fmt.Printf("Found %d weight entries:\n\n", len(entries))
	for _, entry := range entries {
		printWeightEntry(entry)
	}
}
