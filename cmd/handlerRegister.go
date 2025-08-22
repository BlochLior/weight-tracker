package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/BlochLior/weight-tracker/internal/db"
	"github.com/google/uuid"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Fprintf(os.Stderr, "user %s already exists\n", name)
		os.Exit(1)
	}

	user, err := s.db.CreateUser(
		context.Background(), db.CreateUserParams{
			ID:        uuid.NewString(),
			Username:  name,
			CreatedAt: time.Now().Format("02/01/2006"),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	err = s.cfg.SetUser(user.Username)
	if err != nil {
		return fmt.Errorf("failed setting current user %s in config", user.Username)
	}
	fmt.Println("User was created successfuly")
	printUser(user)
	return nil
}
