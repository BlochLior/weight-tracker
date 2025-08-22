package main

import (
	"log"
	"os"

	"github.com/BlochLior/weight-tracker/internal/config"
	"github.com/BlochLior/weight-tracker/internal/db"
	_ "modernc.org/sqlite"
)

// This is the CLI entry point

type state struct {
	db  *db.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error trying to read config: %s", err)
	}

	database, err := db.InitDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}

	dbQueries := db.New(database)

	trackerState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("users", handlerUsersList)
	cmds.register("reset", middlewareLoggedIn(handlerReset))
	cmds.register("weights", middlewareLoggedIn(handlerGetWeights))
	cmds.register("add", middlewareLoggedIn(handlerAdd))
	cmds.register("delete", middlewareLoggedIn(handlerDeleteWeights))
	cmds.register("help", handlerHelp)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
		return
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(trackerState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
