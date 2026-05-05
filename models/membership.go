package models

import "time"

// UserRole represents the role of a user within a tenant.
type UserRole = string

// User role values.
const (
	UserRoleViewer UserRole = "viewer"
	UserRoleEditor UserRole = "editor"
	UserRoleAdmin  UserRole = "admin"
	UserRoleOwner  UserRole = "owner"
)

// Membership represents a user's role within a specific tenant.
// A user can have multiple memberships across different tenants.
type Membership struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()"`
	Role      UserRole  `json:"role" db:"role" constraints:"notnull" default:"'viewer'"`
	ID        string    `json:"id" db:"id" constraints:"primarykey"`
	UserID    string    `json:"user_id" db:"user_id" constraints:"notnull"`
	TenantID  string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
}

// GetID returns the membership's primary key.
func (m Membership) GetID() string {
	return m.ID
}

// GetCreatedAt returns the membership's creation timestamp.
func (m Membership) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// Clone returns a copy of the membership.
func (m Membership) Clone() Membership {
	return m
}
