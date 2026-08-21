package models

import "time"

// Scope is a permission defined by an application. Applications declare their
// own scope catalog; janus stores the definitions but does not interpret them.
// Scopes are bundled into tiers via features, and granted to users on a Grant.
type Scope struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	ApplicationID string    `json:"application_id" db:"application_id" constraints:"notnull"`
	Name          string    `json:"name" db:"name" constraints:"notnull"`
	Description   string    `json:"description" db:"description"`
}

// GetID returns the primary key.
func (s Scope) GetID() string {
	return s.ID
}

// GetCreatedAt returns the creation timestamp.
func (s Scope) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// Clone returns a copy of the scope.
func (s Scope) Clone() Scope {
	return s
}
