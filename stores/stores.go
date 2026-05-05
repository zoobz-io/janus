package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
)

// Stores aggregates all data store instances for the application.
type Stores struct {
	Tenants          *Tenants
	Users            *Users
	Memberships      *Memberships
	LinkedIdentities *LinkedIdentities
	Sessions         *Sessions
}

// New initializes all stores and returns an aggregated Stores instance.
func New(db *sqlx.DB, renderer astql.Renderer) *Stores {
	return &Stores{
		Tenants:          NewTenants(db, renderer),
		Users:            NewUsers(db, renderer),
		Memberships:      NewMemberships(db, renderer),
		LinkedIdentities: NewLinkedIdentities(db, renderer),
		Sessions:         NewSessions(db, renderer),
	}
}
