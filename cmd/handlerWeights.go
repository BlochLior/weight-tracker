package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/BlochLior/weight-tracker/internal/db"
	"github.com/google/uuid"
)

func handlerGetWeights(s *state, cmd command, user db.User) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	err := printWeights(s)
	if err != nil {
		return err
	}
	return nil
}

func printWeights(s *state) error {
	weights, err := s.db.GetAllUserWeights(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	for _, weight := range weights {
		printWeight(weight)
		fmt.Println()
	}
	return nil
}

func printWeight(weight db.Weight) {
	fmt.Println("* Weight: %s %s\n", weight.Weight, weight.WeightUnit)
	fmt.Println("* Entry Date: %s\n", weight.Date)
	if weight.Note.Valid {
		fmt.Println("* Note: %s\n", weight.Note.String)
	}
	fmt.Println("* ID: %s\n", weight.ID)
}

// handlerAdd: adds a weight entry - "dd/mm/yyyy" REALval
func handlerAdd(s *state, cmd command, user db.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: %s <name>: <date:\"dd/mm/yyyy\">, <weight>", cmd.Name)
	}
	date := cmd.Args[0]
	weight := cmd.Args[1]
	note := sql.NullString{
		String: "",
		Valid:  false,
	}
	if len(cmd.Args) > 2 {
		note.Valid = true
		if len(cmd.Args[2:]) > 1 {
			note.String = strings.Join(cmd.Args[2:], " ")
		} else {
			note.String = cmd.Args[2]
		}
	}
	parsedWeight, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		return fmt.Errorf("failed to parse weight '%s': %w", weight, err)
	}

	newWeight, err := s.db.AddWeightEntry(context.Background(), db.AddWeightEntryParams{
		ID:         uuid.NewString(),
		Date:       date,
		Weight:     parsedWeight,
		WeightUnit: "kg",
		Note:       note,
	})
	if err != nil {
		return fmt.Errorf("failed adding new weight entry to database: %w", err)
	}
	err = s.db.AddWeightsUsersEntry(context.Background(), db.AddWeightsUsersEntryParams{
		UserID:   user.ID,
		WeightID: newWeight.ID,
	})
	if err != nil {
		return fmt.Errorf("failed adding to join table in database: %w", err)
	}
	fmt.Println("Successfully added weight entry to database:")
	printWeight(newWeight)
	return nil

}

// Deletes specific weight entry for the current user, given a weight id. Id can be retrieved from the command <weights>.
// If instead of id, given the flag --all, will clear all weight entries for current user.
func handlerDeleteWeights(s *state, cmd command, user db.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: %s <name>: <weight_id OR --all>", cmd.Name)
	}
	arg := struct {
		value   string
		allFlag bool
	}{
		value:   cmd.Args[0],
		allFlag: false,
	}
	if arg.value == "--all" {
		arg.allFlag = true
	}
	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("unexpected error while trying to retrieve current user: %w", err)
	}
	if !arg.allFlag {
		weightEntry, err := s.db.GetWeightEntry(context.Background(), arg.value)
		if err != nil {
			return fmt.Errorf("unable to retrieve weight entry from database, to safeguard delete: %w", err)
		}
		userID, err := s.db.GetUserIDFromWeightID(context.Background(), arg.value)
		if err != nil {
			return fmt.Errorf("provided %s isn't a supported flag or doesn't exist in database: %w", arg.value, err)
		}
		if currentUser.ID != userID {
			return fmt.Errorf("prohibited deletion of records: weight of id %s is not related to %s", arg.value, s.cfg.CurrentUserName)
		}
		err = s.db.DeleteUserWeight(context.Background(), arg.value)
		if err != nil {
			return fmt.Errorf("unable to delete weight record: %w", err)
		}
		fmt.Println("Successfully deleted weight entry:")
		printWeight(weightEntry)
		return nil
	} else {
		// this covers the true value for the arg.allFlag
		err := s.db.DeleteAllUserWeights(context.Background(), currentUser.ID)
		if err != nil {
			return fmt.Errorf("failed to delete all user weights: %w", err)
		}
		fmt.Printf("Successfully deleted all weight values for %s", s.cfg.CurrentUserName)
		return nil
	}
}

// time.Now().Format("02/01/2006")
