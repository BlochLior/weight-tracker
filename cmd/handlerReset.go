package main

import (
	"context"
	"fmt"

	"github.com/BlochLior/weight-tracker/internal/db"
)

// Functionally, this is probably not a needed command, but i added it anyways, together with a middleware
func handlerReset(s *state, cmd command, user db.User) error {
	err := s.db.DeleteUsersData(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}
	err = s.db.DeleteWeightData(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete weights: %w", err)
	}
	// to finish and make sure that this works, i need to explore how foreign keys work with sqlite and apply them in the add_users_weights_table.sql
	fmt.Println("Database reset completed successfully.")
	return nil
}
