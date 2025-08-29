package tracker

import (
	"database/sql"
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
