package models

import "time"

// License records that a tenant is authorized to use
// a specific application. Created by the system when a tenant
// is granted access to an app.
type License struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID      string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ApplicationID string    `json:"application_id" db:"application_id" constraints:"notnull"`
}

// GetID returns the primary key.
func (ta License) GetID() string {
	return ta.ID
}

// GetCreatedAt returns the creation timestamp.
func (ta License) GetCreatedAt() time.Time {
	return ta.CreatedAt
}

// Clone returns a copy.
func (ta License) Clone() License {
	return ta
}
