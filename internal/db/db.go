package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// OpenDB loads the database path from .env and returns a database connection
func OpenDB() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DATABASE_PATH not set in .env file")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	return db, nil
}
