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

// UserStatus represents the account status of a user.
type UserStatus = string

// User status values.
const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

// User represents an authenticated person within a tenant.
type User struct {
	CreatedAt   time.Time  `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" default:"now()"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	Role        UserRole   `json:"role" db:"role" constraints:"notnull" default:"'viewer'"`
	Status      UserStatus `json:"status" db:"status" constraints:"notnull" default:"'active'"`
	ID          string     `json:"id" db:"id" constraints:"primarykey"`
	TenantID    string     `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	Email       string     `json:"email" db:"email" constraints:"notnull"`
	DisplayName string     `json:"display_name" db:"display_name" constraints:"notnull"`
}

// GetID returns the user's primary key.
func (u User) GetID() string {
	return u.ID
}

// GetCreatedAt returns the user's creation timestamp.
func (u User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// Clone returns a deep copy of the user.
func (u User) Clone() User {
	c := u
	if u.LastSeenAt != nil {
		t := *u.LastSeenAt
		c.LastSeenAt = &t
	}
	return c
}
