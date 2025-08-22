package main

import (
	"fmt"
)

// i am using some guidance from the gator project
type command struct {
	Name string
	Args []string
}
type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("command %s not found", cmd.Name)
	}
	return handler(s, cmd)
}
