package tracker

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a weight entry by ID",
	Long: `Delete a weight entry from the database by its ID.

You can find the ID of entries by using the 'list' command.

Examples:
  weight-tracker delete 1                    # Delete entry with ID 1
  weight-tracker delete 5 --confirm          # Delete entry with ID 5 (skip confirmation)
  weight-tracker delete 3 --force            # Force delete without confirmation
`,
	Args: cobra.ExactArgs(1),
	Run:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")
}

// runDeleteInternal contains the core logic and returns errors instead of terminating
func runDeleteInternal(cmd *cobra.Command, args []string) error {
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

	// Check if entry exists before deletion
	existingEntry, err := store.GetWeight(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to find weight entry with ID %d: %w", id, err)
	}

	// Show what will be deleted
	fmt.Printf("Found weight entry to delete:\n")
	printWeightEntry(existingEntry)

	// Handle confirmation unless --confirm or --force is used
	confirmDelete, _ := cmd.Flags().GetBool("confirm")
	forceDelete, _ := cmd.Flags().GetBool("force")

	if !confirmDelete && !forceDelete {
		fmt.Print("Are you sure you want to delete this entry? (y/N): ")
		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the entry
	err = store.DeleteWeight(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to delete weight entry: %w", err)
	}

	fmt.Printf("Successfully deleted weight entry with ID %d.\n", id)
	return nil
}

// runDelete is the cobra command wrapper that handles errors appropriately for CLI usage
func runDelete(cmd *cobra.Command, args []string) {
	if err := runDeleteInternal(cmd, args); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
