package models

import "time"

// Tier is a subscription level defined by an application. Each tier bundles a
// set of scopes via features. A user's Grant may be placed on a tier.
type Tier struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	ApplicationID string    `json:"application_id" db:"application_id" constraints:"notnull"`
	Slug          string    `json:"slug" db:"slug" constraints:"notnull"`
	Name          string    `json:"name" db:"name" constraints:"notnull"`
	Rank          int       `json:"rank" db:"rank" default:"0"`
}

// GetID returns the primary key.
func (t Tier) GetID() string {
	return t.ID
}

// GetCreatedAt returns the creation timestamp.
func (t Tier) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// Clone returns a copy of the tier.
func (t Tier) Clone() Tier {
	return t
}
