package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/BlochLior/weight-tracker/internal/db"
	"github.com/BlochLior/weight-tracker/internal/db/sqlc"
)

// WeightEntry represents a weight entry in our application domain
// This is the struct that commands work with for validation and business logic
type WeightEntry struct {
	ID     int64     `json:"id"`
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date"`
	Unit   string    `json:"unit"`
	Note   string    `json:"note"`
	UserID string    `json:"user_id"`
}

// ListOptions represents filtering and sorting options for listing weight entries
type ListOptions struct {
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	SortBy   string     `json:"sort_by,omitempty"` // "date" or "weight"
	SortDesc bool       `json:"sort_desc,omitempty"`
	Unit     string     `json:"unit,omitempty"`
}

// Store defines the contract for weight entry storage operations
// This interface allows for easy testing with mock implementations
type Store interface {
	// AddWeight adds a new weight entry to the store
	AddWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error)

	// ListWeights retrieves weight entries based on the provided options
	ListWeights(ctx context.Context, options ListOptions) ([]WeightEntry, error)

	// DeleteWeight removes a weight entry by ID
	DeleteWeight(ctx context.Context, id int64) error

	// GetWeight retrieves a single weight entry by ID
	GetWeight(ctx context.Context, id int64) (WeightEntry, error)

	// UpdateWeight updates an existing weight entry
	UpdateWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error)
}

// DBStore is the concrete implementation of Store that uses SQLite and sqlc
type DBStore struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// NewDBStore creates a new DBStore instance
func NewDBStore() (*DBStore, error) {
	db, err := db.OpenDB()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	queries := sqlc.New(db)

	return &DBStore{
		db:      db,
		queries: queries,
	}, nil
}

// NewDBStoreWithDB creates a new DBStore instance with an existing database connection
// This is useful for testing with in-memory databases
func NewDBStoreWithDB(database *sql.DB) *DBStore {
	queries := sqlc.New(database)
	return &DBStore{
		db:      database,
		queries: queries,
	}
}

// Close closes the database connection
func (s *DBStore) Close() error {
	return s.db.Close()
}

// AddWeight adds a new weight entry to the database
func (s *DBStore) AddWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error) {
	// Convert WeightEntry to sqlc format
	params := sqlc.AddWeightParams{
		Weight: entry.Weight,
		Date:   sql.NullString{String: FormatDateForDB(entry.Date), Valid: true},
		Unit:   sql.NullString{String: entry.Unit, Valid: entry.Unit != ""},
		Note:   sql.NullString{String: entry.Note, Valid: entry.Note != ""},
	}

	// Call sqlc method
	sqlcEntry, err := s.queries.AddWeight(ctx, params)
	if err != nil {
		return WeightEntry{}, fmt.Errorf("failed to add weight entry: %w", err)
	}

	// Convert back to WeightEntry
	return s.sqlcToWeightEntry(sqlcEntry), nil
}

// ListWeights retrieves weight entries based on the provided options
func (s *DBStore) ListWeights(ctx context.Context, options ListOptions) ([]WeightEntry, error) {
	// Set defaults
	if options.SortBy == "" {
		options.SortBy = "date"
	}
	if options.Limit == 0 {
		options.Limit = -1 // SQLite treats -1 as no limit
	}

	// Convert date filters to interface{} (as expected by sqlc)
	var startDate, endDate interface{}
	if options.FromDate != nil {
		startDate = FormatDateForDB(*options.FromDate)
	}
	if options.ToDate != nil {
		endDate = FormatDateForDB(*options.ToDate)
	}

	// Call appropriate sqlc method based on sort options
	var sqlcEntries []sqlc.Weight
	var err error

	switch {
	case options.SortBy == "date" && options.SortDesc:
		params := sqlc.ListWeightsDateDescParams{
			StartDate: startDate,
			EndDate:   endDate,
			RowLimit:  int64(options.Limit),
		}
		sqlcEntries, err = s.queries.ListWeightsDateDesc(ctx, params)

	case options.SortBy == "date" && !options.SortDesc:
		params := sqlc.ListWeightsDateAscParams{
			StartDate: startDate,
			EndDate:   endDate,
			RowLimit:  int64(options.Limit),
		}
		sqlcEntries, err = s.queries.ListWeightsDateAsc(ctx, params)

	case options.SortBy == "weight" && options.SortDesc:
		params := sqlc.ListWeightsWeightDescParams{
			StartDate: startDate,
			EndDate:   endDate,
			RowLimit:  int64(options.Limit),
		}
		sqlcEntries, err = s.queries.ListWeightsWeightDesc(ctx, params)

	case options.SortBy == "weight" && !options.SortDesc:
		params := sqlc.ListWeightsWeightAscParams{
			StartDate: startDate,
			EndDate:   endDate,
			RowLimit:  int64(options.Limit),
		}
		sqlcEntries, err = s.queries.ListWeightsWeightAsc(ctx, params)

	default:
		// Default to date desc
		params := sqlc.ListWeightsDateDescParams{
			StartDate: startDate,
			EndDate:   endDate,
			RowLimit:  int64(options.Limit),
		}
		sqlcEntries, err = s.queries.ListWeightsDateDesc(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list weight entries: %w", err)
	}

	// Convert sqlc entries to WeightEntry
	entries := make([]WeightEntry, len(sqlcEntries))
	for i, sqlcEntry := range sqlcEntries {
		entries[i] = s.sqlcToWeightEntry(sqlcEntry)
	}

	// Apply unit filtering (not supported by current sqlc queries)
	if options.Unit != "" {
		filtered := make([]WeightEntry, 0)
		for _, entry := range entries {
			if entry.Unit == options.Unit {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	return entries, nil
}

// DeleteWeight removes a weight entry by ID
func (s *DBStore) DeleteWeight(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	// Check if the entry exists first
	_, err := s.queries.GetWeight(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("weight entry with id %d not found", id)
		}
		return fmt.Errorf("failed to check if weight entry exists: %w", err)
	}

	// Delete the entry
	err = s.queries.DeleteWeight(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete weight entry: %w", err)
	}
	return nil
}

// GetWeight retrieves a single weight entry by ID
func (s *DBStore) GetWeight(ctx context.Context, id int64) (WeightEntry, error) {
	sqlcEntry, err := s.queries.GetWeight(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return WeightEntry{}, fmt.Errorf("weight entry with id %d not found", id)
		}
		return WeightEntry{}, fmt.Errorf("failed to get weight entry: %w", err)
	}

	return s.sqlcToWeightEntry(sqlcEntry), nil
}

// UpdateWeight updates an existing weight entry
func (s *DBStore) UpdateWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error) {
	if entry.ID <= 0 {
		return WeightEntry{}, fmt.Errorf("invalid ID: %d", entry.ID)
	}

	// Get the existing entry first
	existingEntry, err := s.GetWeight(ctx, entry.ID)
	if err != nil {
		return WeightEntry{}, err
	}

	// Merge the update with existing entry (partial update)
	updatedEntry := existingEntry

	// Only update fields that are not zero values
	if entry.Weight != 0 {
		updatedEntry.Weight = entry.Weight
	}
	if !entry.Date.IsZero() {
		updatedEntry.Date = entry.Date
	}
	if entry.Unit != "" {
		updatedEntry.Unit = entry.Unit
	}
	if entry.Note != "" {
		updatedEntry.Note = entry.Note
	}
	if entry.UserID != "" {
		updatedEntry.UserID = entry.UserID
	}

	// Validate the merged entry
	if err := ValidateWeightEntry(updatedEntry); err != nil {
		return WeightEntry{}, err
	}

	// Convert WeightEntry to sqlc format
	params := sqlc.UpdateWeightParams{
		Weight: updatedEntry.Weight,
		Date:   sql.NullString{String: FormatDateForDB(updatedEntry.Date), Valid: true},
		Unit:   sql.NullString{String: updatedEntry.Unit, Valid: updatedEntry.Unit != ""},
		Note:   sql.NullString{String: updatedEntry.Note, Valid: updatedEntry.Note != ""},
		UserID: sql.NullString{String: updatedEntry.UserID, Valid: updatedEntry.UserID != ""},
		ID:     updatedEntry.ID,
	}

	// Call sqlc method
	sqlcEntry, err := s.queries.UpdateWeight(ctx, params)
	if err != nil {
		return WeightEntry{}, fmt.Errorf("failed to update weight entry: %w", err)
	}

	// Convert back to WeightEntry
	return s.sqlcToWeightEntry(sqlcEntry), nil
}

// sqlcToWeightEntry converts a sqlc.Weight to WeightEntry
func (s *DBStore) sqlcToWeightEntry(sqlcEntry sqlc.Weight) WeightEntry {
	entry := WeightEntry{
		ID:     sqlcEntry.ID,
		Weight: sqlcEntry.Weight,
	}

	// Parse date (always stored in ISO format in database)
	if sqlcEntry.Date.Valid && sqlcEntry.Date.String != "" {
		if date, err := time.Parse(DBFormat, sqlcEntry.Date.String); err == nil {
			entry.Date = date
		} else {
			// Fallback to zero time if parsing fails
			entry.Date = time.Time{}
		}
	} else {
		entry.Date = time.Time{}
	}

	// Handle optional fields
	if sqlcEntry.Unit.Valid {
		entry.Unit = sqlcEntry.Unit.String
	}

	if sqlcEntry.Note.Valid {
		entry.Note = sqlcEntry.Note.String
	}

	if sqlcEntry.UserID.Valid {
		entry.UserID = sqlcEntry.UserID.String
	}

	return entry
}

// ValidateWeightEntry validates a WeightEntry struct
func ValidateWeightEntry(entry WeightEntry) error {
	if entry.Weight <= 0 {
		return fmt.Errorf("weight must be greater than 0")
	}

	if entry.Unit != "" && entry.Unit != "kg" && entry.Unit != "lbs" {
		return fmt.Errorf("unit must be 'kg' or 'lbs', got: %s", entry.Unit)
	}

	return nil
}
