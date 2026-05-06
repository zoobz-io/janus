package models

import "time"

// UserApplication records that a user has been granted access to
// an application within a specific tenant. The tenant must already
// be authorized for the application (TenantApplication).
type UserApplication struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	UserID        string    `json:"user_id" db:"user_id" constraints:"notnull"`
	TenantID      string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ApplicationID string    `json:"application_id" db:"application_id" constraints:"notnull"`
}

// GetID returns the primary key.
func (ua UserApplication) GetID() string {
	return ua.ID
}

// GetCreatedAt returns the creation timestamp.
func (ua UserApplication) GetCreatedAt() time.Time {
	return ua.CreatedAt
}

// Clone returns a copy.
func (ua UserApplication) Clone() UserApplication {
	return ua
}
