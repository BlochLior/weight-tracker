package main

import (
	"context"
	"fmt"
	"os"

	"github.com/BlochLior/weight-tracker/internal/db"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user db.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Printf("given %s username doesn't exist. %s", name, err)
		os.Exit(1)
		return err
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfuly!")
	return nil
}

func handlerUsersList(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list users: %w", err)
	}
	for _, user := range users {
		if user.Username == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Username)
			continue
		}
		fmt.Printf("* %v\n", user.Username)
	}
	return nil
}

func printUser(user db.User) {
	fmt.Printf(" * User ID: %v\n", user.ID)
	fmt.Printf(" * Username: %v\n", user.Username)
}
