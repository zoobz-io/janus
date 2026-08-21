package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
)

// Stores aggregates all data store instances for the application.
type Stores struct {
	db           *sqlx.DB
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
		db:           db,
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

// WithTx runs fn inside a single database transaction: commit on nil return,
// rollback on error. Store methods with Tx variants participate by taking the
// callback's tx. Multi-write invariants (a tenant always has an owner, a
// registered user always has a linked identity) depend on this.
func (s *Stores) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rolling back: %w", rbErr))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
