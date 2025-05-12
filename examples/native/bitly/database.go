package bitly

import (
	sdlc "github.com/panyam/sdl/components"
	sdl "github.com/panyam/sdl/core"
)

// DatabaseComponent encapsulates storage logic for Bitly.
// Uses an underlying Index primitive (e.g., HashIndex or BTreeIndex).
type DatabaseComponent struct {
	Name string
	// The primary index mapping shortCode -> longURL
	PrimaryIndex sdlc.HashIndex
	// Optional: Index for longURL -> shortCode (for checking existence)
	// SecondaryIndex Index

	// Configuration
	MaxOutcomeLen int
}

// Init initializes the DatabaseComponent.
// Requires specifying the index type to use.
func (db *DatabaseComponent) Init(name string) *DatabaseComponent {
	db.Name = name
	db.MaxOutcomeLen = 10 // Default

	// Create and initialize the underlying index
	// TODO: Add factory pattern or more flexible index selection
	db.PrimaryIndex.MaxOutcomeLen = db.MaxOutcomeLen // Propagate setting
	db.PrimaryIndex.Init()

	// TODO: Initialize SecondaryIndex if needed

	return db
}

// GetLongURL retrieves the long URL for a given short code.
func (db *DatabaseComponent) GetLongURL(shortCode string) *sdl.Outcomes[sdl.AccessResult] {
	// Check the type of index and call its Find method
	return db.PrimaryIndex.Find()
}

// SaveMapping stores the shortCode -> longURL mapping.
func (db *DatabaseComponent) SaveMapping(shortCode, longUrl string) *sdl.Outcomes[sdl.AccessResult] {
	return db.PrimaryIndex.Insert()
}

// TODO: Add GetShortCode(longURL) if needed
