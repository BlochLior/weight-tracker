package tracker

import (
	"context"
	"fmt"
)

// store_mock.go - MockStore implementation
// * purpose: contains the MockStore struct and methods
// * content: The actual MockStore implementation that implements the Store interface
// * focus: in-memory storage for testing purposes

// MockStore is an in-memory implementation of the Store interface for testing
type MockStore struct {
	entries []WeightEntry
	nextID  int64
}

// NewMockStore creates a new MockStore instance
func NewMockStore() *MockStore {
	return &MockStore{
		entries: make([]WeightEntry, 0),
		nextID:  1,
	}
}

// AddWeight adds a new weight entry to the mock store
func (m *MockStore) AddWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error) {
	entry.ID = m.nextID
	m.nextID++
	m.entries = append(m.entries, entry)
	return entry, nil
}

// ListWeights retrieves weight entries from the mock store
func (m *MockStore) ListWeights(ctx context.Context, options ListOptions) ([]WeightEntry, error) {
	// Start with all entries
	result := make([]WeightEntry, len(m.entries))
	copy(result, m.entries)

	// Apply date filtering
	if options.FromDate != nil || options.ToDate != nil {
		filtered := make([]WeightEntry, 0)
		for _, entry := range result {
			include := true

			if options.FromDate != nil && entry.Date.Before(*options.FromDate) {
				include = false
			}
			if options.ToDate != nil && entry.Date.After(*options.ToDate) {
				include = false
			}

			if include {
				filtered = append(filtered, entry)
			}
		}
		result = filtered
	}

	// Apply unit filtering
	if options.Unit != "" {
		filtered := make([]WeightEntry, 0)
		for _, entry := range result {
			if entry.Unit == options.Unit {
				filtered = append(filtered, entry)
			}
		}
		result = filtered
	}

	// Apply sorting
	if options.SortBy != "" {
		switch options.SortBy {
		case "date":
			if options.SortDesc {
				// Sort by date descending
				for i := 0; i < len(result)-1; i++ {
					for j := i + 1; j < len(result); j++ {
						if result[i].Date.Before(result[j].Date) {
							result[i], result[j] = result[j], result[i]
						}
					}
				}
			} else {
				// Sort by date ascending
				for i := 0; i < len(result)-1; i++ {
					for j := i + 1; j < len(result); j++ {
						if result[i].Date.After(result[j].Date) {
							result[i], result[j] = result[j], result[i]
						}
					}
				}
			}
		case "weight":
			if options.SortDesc {
				// Sort by weight descending
				for i := 0; i < len(result)-1; i++ {
					for j := i + 1; j < len(result); j++ {
						if result[i].Weight < result[j].Weight {
							result[i], result[j] = result[j], result[i]
						}
					}
				}
			} else {
				// Sort by weight ascending
				for i := 0; i < len(result)-1; i++ {
					for j := i + 1; j < len(result); j++ {
						if result[i].Weight > result[j].Weight {
							result[i], result[j] = result[j], result[i]
						}
					}
				}
			}
		}
	}

	// Apply limit
	if options.Limit > 0 && options.Limit < len(result) {
		result = result[:options.Limit]
	}

	return result, nil
}

// DeleteWeight removes a weight entry by ID from the mock store
func (m *MockStore) DeleteWeight(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	for i, entry := range m.entries {
		if entry.ID == id {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("weight entry with id %d not found", id)
}

// GetWeight retrieves a single weight entry by ID from the mock store
func (m *MockStore) GetWeight(ctx context.Context, id int64) (WeightEntry, error) {
	for _, entry := range m.entries {
		if entry.ID == id {
			return entry, nil
		}
	}
	return WeightEntry{}, fmt.Errorf("weight entry with id %d not found", id)
}

// UpdateWeight updates an existing weight entry in the mock store
func (m *MockStore) UpdateWeight(ctx context.Context, entry WeightEntry) (WeightEntry, error) {
	if entry.ID <= 0 {
		return WeightEntry{}, fmt.Errorf("invalid ID: %d", entry.ID)
	}

	// Validate the entry
	if err := ValidateWeightEntry(entry); err != nil {
		return WeightEntry{}, err
	}

	for i, existingEntry := range m.entries {
		if existingEntry.ID == entry.ID {
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

			m.entries[i] = updatedEntry
			return updatedEntry, nil
		}
	}
	return WeightEntry{}, fmt.Errorf("weight entry with id %d not found", entry.ID)
}

// Close is a no-op for the mock store
func (m *MockStore) Close() error {
	return nil
}
