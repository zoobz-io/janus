package models

import (
	"time"

	"github.com/lib/pq"
)

// UserApplication records that a user has been granted access to
// an application within a specific tenant, along with application-specific
// roles and scopes. The tenant must already be authorized for the
// application (TenantApplication).
//
// Roles and Scopes are defined by each application — janus stores
// and returns them but does not interpret them. They are returned
// in the AuthorizedTenant response during session validation so the
// calling app can authorize without extra round-trips.
type UserApplication struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	UserID        string    `json:"user_id" db:"user_id" constraints:"notnull"`
	TenantID      string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ApplicationID string    `json:"application_id" db:"application_id" constraints:"notnull"`
	Roles         pq.StringArray `json:"roles" db:"roles" type:"text[]"`
	Scopes        pq.StringArray `json:"scopes" db:"scopes" type:"text[]"`
}

// GetID returns the primary key.
func (ua UserApplication) GetID() string {
	return ua.ID
}

// GetCreatedAt returns the creation timestamp.
func (ua UserApplication) GetCreatedAt() time.Time {
	return ua.CreatedAt
}

// Clone returns a deep copy.
func (ua UserApplication) Clone() UserApplication {
	c := ua
	if ua.Roles != nil {
		c.Roles = make([]string, len(ua.Roles))
		copy(c.Roles, ua.Roles)
	}
	if ua.Scopes != nil {
		c.Scopes = make([]string, len(ua.Scopes))
		copy(c.Scopes, ua.Scopes)
	}
	return c
}
