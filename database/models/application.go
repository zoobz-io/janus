package models

import "time"

// ApplicationStatus represents the lifecycle status of an application.
type ApplicationStatus = string

// Application status values.
const (
	ApplicationStatusActive   ApplicationStatus = "active"
	ApplicationStatusInactive ApplicationStatus = "inactive"
)

// Application represents a product in the zoobz-io ecosystem.
// Each application is a service on the mesh — the Slug matches
// the cert CN used by that service's nodes. Applications are
// pre-registered by the system; tenants are authorized to use
// them, and can delegate access to individual users.
type Application struct {
	CreatedAt time.Time         `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at" default:"now()"`
	Status    ApplicationStatus `json:"status" db:"status" constraints:"notnull" default:"'active'"`
	Name      string            `json:"name" db:"name" constraints:"notnull"`
	Slug      string            `json:"slug" db:"slug" constraints:"notnull,unique"`
	ID        string            `json:"id" db:"id" constraints:"primarykey"`
}

// GetID returns the application's primary key.
func (a Application) GetID() string {
	return a.ID
}

// GetCreatedAt returns the application's creation timestamp.
func (a Application) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// Clone returns a copy of the application.
func (a Application) Clone() Application {
	return a
}
