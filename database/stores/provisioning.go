package stores

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/zoobz-io/janus/database/models"
)

// Provisioning operations: the multi-store write flows whose invariants only
// hold if every write lands or none do. These are the only sanctioned ways to
// perform these flows — composing the individual store methods at a call site
// reintroduces the partial-write windows these exist to close.

// CreateTenantWithOwner creates a tenant and its owner membership atomically.
// A tenant must never exist without an owner — the owner-protection rules in
// authz assume it.
func (s *Stores) CreateTenantWithOwner(ctx context.Context, name, slug, ownerUserID string) (*models.Tenant, *models.Membership, error) {
	var tenant *models.Tenant
	var membership *models.Membership
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		var txErr error
		tenant, txErr = s.Tenants.CreateTenantTx(ctx, tx, name, slug)
		if txErr != nil {
			return txErr
		}
		membership, txErr = s.Memberships.CreateTx(ctx, tx, ownerUserID, tenant.ID, models.UserRoleOwner)
		return txErr
	})
	if err != nil {
		return nil, nil, err
	}
	return tenant, membership, nil
}

// RegisterUser creates a user and links the external identity atomically, and,
// when tenantName and tenantSlug are both set, creates an owned tenant in the
// same transaction. A registered user must never exist without a linked
// identity — without one there is no way to log in and reach the record.
// The returned tenant is nil when no tenant was requested.
func (s *Stores) RegisterUser(ctx context.Context, email, displayName, provider, externalSubject, tenantName, tenantSlug string) (*models.User, *models.Tenant, error) {
	var user *models.User
	var tenant *models.Tenant
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		var txErr error
		user, txErr = s.Users.CreateUserTx(ctx, tx, email, displayName)
		if txErr != nil {
			return txErr
		}
		if _, txErr = s.Accounts.LinkTx(ctx, tx, user.ID, provider, externalSubject); txErr != nil {
			return txErr
		}
		if tenantName != "" && tenantSlug != "" {
			tenant, txErr = s.Tenants.CreateTenantTx(ctx, tx, tenantName, tenantSlug)
			if txErr != nil {
				return txErr
			}
			if _, txErr = s.Memberships.CreateTx(ctx, tx, user.ID, tenant.ID, models.UserRoleOwner); txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return user, tenant, nil
}
