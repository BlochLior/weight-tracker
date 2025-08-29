-- +goose Up
-- Rename the old table
ALTER TABLE weights RENAME TO old_weights;

-- Create the new table with nullable columns and default values
CREATE TABLE weights (
    id INTEGER PRIMARY KEY,
    weight REAL NOT NULL,
    date TEXT DEFAULT CURRENT_TIMESTAMP,
    unit TEXT DEFAULT 'kg',
    note TEXT,
    user_id TEXT
);

-- Copy data from the old table to the new one
-- This command will copy the old NOT NULL values over.
INSERT INTO weights (id, weight, date, unit, note, user_id)
SELECT id, weight, date, unit, note, user_id FROM old_weights;

-- Drop the old table
DROP TABLE old_weights;

-- +goose Down
-- Re-create the old table
CREATE TABLE weights (
    id INTEGER PRIMARY KEY,
    weight REAL NOT NULL,
    date TEXT NOT NULL,
    unit TEXT NOT NULL,
    note TEXT,
    user_id TEXT
);

-- Copy data back (existing data will have non-null values)
INSERT INTO weights (id, weight, date, unit, note, user_id)
SELECT id, weight, date, unit, note, user_id FROM old_weights;

-- Drop the new (current) table
DROP TABLE old_weights;