package main

import (
	"github.com/BlochLior/weight-tracker/cmd/tracker"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Application entry point for Cobra apps
	tracker.Execute()
}
