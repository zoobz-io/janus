// Package models defines the core domain entities for janus.
package models

import "time"

// UserStatus represents the account status of a user.
type UserStatus = string

// User status values.
const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

// User represents an authenticated person in the system.
// Users can belong to multiple tenants via Memberships.
type User struct {
	CreatedAt   time.Time  `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" default:"now()"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	Status      UserStatus `json:"status" db:"status" constraints:"notnull" default:"'active'"`
	ID          string     `json:"id" db:"id" constraints:"primarykey"`
	Email       string     `json:"email" db:"email" constraints:"notnull,unique"`
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
