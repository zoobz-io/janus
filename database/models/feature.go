package models

import "time"

// Feature bundles a scope into a tier — one row records that a tier includes
// a given scope. Deleting either the tier or the scope cascades to its features.
type Feature struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()"`
	ID        string    `json:"id" db:"id" constraints:"primarykey"`
	TierID    string    `json:"tier_id" db:"tier_id" constraints:"notnull"`
	ScopeID   string    `json:"scope_id" db:"scope_id" constraints:"notnull"`
}

// GetID returns the primary key.
func (f Feature) GetID() string {
	return f.ID
}

// GetCreatedAt returns the creation timestamp.
func (f Feature) GetCreatedAt() time.Time {
	return f.CreatedAt
}

// Clone returns a copy of the feature.
func (f Feature) Clone() Feature {
	return f
}
