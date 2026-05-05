package models

import "time"

// TenantStatus represents the account status of a tenant.
type TenantStatus = string

// Tenant status values.
const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

// Tenant represents a company that has contracted with zoobzio.
// All users belong to exactly one tenant. Downstream apps resolve
// user access through the tenant boundary.
type Tenant struct {
	CreatedAt time.Time    `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at" default:"now()"`
	Status    TenantStatus `json:"status" db:"status" constraints:"notnull" default:"'active'"`
	Name      string       `json:"name" db:"name" constraints:"notnull"`
	Slug      string       `json:"slug" db:"slug" constraints:"notnull,unique"`
	ID        string       `json:"id" db:"id" constraints:"primarykey"`
}

// GetID returns the tenant's primary key.
func (t Tenant) GetID() string {
	return t.ID
}

// GetCreatedAt returns the tenant's creation timestamp.
func (t Tenant) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// Clone returns a copy of the tenant.
func (t Tenant) Clone() Tenant {
	return t
}
