package db

import (
	"database/sql"
	"fmt"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// ping database to ensure connection established pre-sending PRAGMA
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// execute PRAGMA foreign_keys = ON for this connection
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close() // closing the connection if the pragma failed
		return nil, fmt.Errorf("failed to enable foreign key support: %w", err)
	}

	fmt.Println("SQLite foreign key enforcement enabled successfully.")
	return db, nil
}
