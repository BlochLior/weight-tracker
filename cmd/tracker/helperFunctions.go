package tracker

import (
	"fmt"

	"github.com/BlochLior/weight-tracker/internal/db/sqlc"
)

func printWeight(weightEntry sqlc.Weight) {
	weightId := weightEntry.ID
	weightValue := weightEntry.Weight
	entryDate := ""
	entryUnit := ""
	entryNote := ""
	entryUser := ""

	if weightEntry.UserID.Valid {
		entryUser = weightEntry.UserID.String
	}
	if weightEntry.Note.Valid {
		entryNote = weightEntry.Note.String
	}
	if weightEntry.Unit.Valid {
		entryUnit = weightEntry.Unit.String
	}
	if weightEntry.Date.Valid {
		entryDate = weightEntry.Date.String
	}

	fmt.Printf("* Weight Entry ID: %d\n", weightId)
	if entryDate != "" { // This should always activate, as date should default if not provided
		fmt.Printf("* Date: %s\n", entryDate)
	}
	if entryUnit != "" { // Same, should default if not provided
		fmt.Printf("* Weight: %.2f %s\n", weightValue, entryUnit)
	}
	if entryNote != "" {
		fmt.Printf("* Note: %s\n", entryNote)
	}
	if entryUser != "" {
		fmt.Printf("* UserID: %s\n", entryUser)
	}

}
