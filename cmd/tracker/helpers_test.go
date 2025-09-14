package tracker

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// setupTestDB creates an in-memory SQLite database and runs all migrations.
// This helper can be used by all command tests to ensure consistent test database setup.
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Get the path to migrations directory
	// Assuming we're in cmd/tracker/, we need to go up two levels to get to migrations/
	migrationsDir := filepath.Join("..", "..", "migrations")

	// Set goose dialect and run migrations
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// failedEntryCreationString returns a formatted error message for test entry creation failures.
// This helper is used across multiple test files to maintain consistent error messaging.
func failedEntryCreationString(err error) string {
	return fmt.Sprintf("Failed to create test entry: %v", err)
}

// failedTestEntryAdditionString returns a formatted error message for test entry addition failures.
// This helper is used across multiple test files to maintain consistent error messaging.
func failedTestEntryAdditionString(err error) string {
	return fmt.Sprintf("Failed to add test entry: %v", err)
}

// unexpectedErrorString returns a formatted error message for unexpected errors in tests.
// This helper is used across multiple test files to maintain consistent error messaging.
func unexpectedErrorString(err error) string {
	return fmt.Sprintf("unexpected error: %v", err)
}
