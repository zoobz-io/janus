package models

import (
	"time"

	"github.com/lib/pq"
)

// Grant records that a user has been granted access to
// an application within a specific tenant, along with application-specific
// roles and scopes. The tenant must already be authorized for the
// application (License).
//
// Roles and Scopes are defined by each application — janus stores
// and returns them but does not interpret them. They are returned
// in the AuthorizedTenant response during session validation so the
// calling app can authorize without extra round-trips.
//
// TierID optionally places the grant on one of the application's tiers;
// the tier must belong to the same application (enforced at the service layer).
type Grant struct {
	CreatedAt     time.Time      `json:"created_at" db:"created_at" default:"now()"`
	TierID        *string        `json:"tier_id,omitempty" db:"tier_id"`
	ID            string         `json:"id" db:"id" constraints:"primarykey"`
	UserID        string         `json:"user_id" db:"user_id" constraints:"notnull"`
	TenantID      string         `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ApplicationID string         `json:"application_id" db:"application_id" constraints:"notnull"`
	Roles         pq.StringArray `json:"roles" db:"roles" type:"text[]"`
	Scopes        pq.StringArray `json:"scopes" db:"scopes" type:"text[]"`
}

// GetID returns the primary key.
func (ua Grant) GetID() string {
	return ua.ID
}

// GetCreatedAt returns the creation timestamp.
func (ua Grant) GetCreatedAt() time.Time {
	return ua.CreatedAt
}

// Clone returns a deep copy.
func (ua Grant) Clone() Grant {
	c := ua
	if ua.TierID != nil {
		t := *ua.TierID
		c.TierID = &t
	}
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
