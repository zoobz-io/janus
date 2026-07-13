package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
)

// Stores aggregates all data store instances for the application.
type Stores struct {
	Tenants      *Tenants
	Users        *Users
	Memberships  *Memberships
	Accounts     *Accounts
	Sessions     *Sessions
	Applications *Applications
	Licenses     *Licenses
	Grants       *Grants
	Scopes       *Scopes
	Tiers        *Tiers
	Features     *Features
	Config       *Config
}

// New initializes all stores and returns an aggregated Stores instance.
func New(db *sqlx.DB, renderer astql.Renderer) *Stores {
	return &Stores{
		Tenants:      NewTenants(db, renderer),
		Users:        NewUsers(db, renderer),
		Memberships:  NewMemberships(db, renderer),
		Accounts:     NewAccounts(db, renderer),
		Sessions:     NewSessions(db, renderer),
		Applications: NewApplications(db, renderer),
		Licenses:     NewLicenses(db, renderer),
		Grants:       NewGrants(db, renderer),
		Scopes:       NewScopes(db, renderer),
		Tiers:        NewTiers(db, renderer),
		Features:     NewFeatures(db, renderer),
		Config:       NewConfig(db, renderer),
	}
}
