package authz

import (
	"context"
	"fmt"

	"github.com/zoobz-io/janus/database/models"
)

// Memberships is the membership access authz needs. Both the public and admin
// membership contracts, and the concrete store, satisfy it structurally — so
// authz stays decoupled from any single API's contract package.
type Memberships interface {
	GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*models.Membership, error)
	CountOtherOwners(ctx context.Context, tenantID, excludeUserID string) (int64, error)
}

// RequireRole verifies the user has one of the given roles in the tenant.
// Returns the membership on success, or ErrNotMember / ErrInsufficientRole.
// A store failure is returned as-is — an infrastructure error must not
// masquerade as an authorization denial.
func RequireRole(ctx context.Context, memberships Memberships, userID, tenantID string, roles ...models.UserRole) (*models.Membership, error) {
	mem, err := memberships.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("looking up membership: %w", err)
	}
	if mem == nil {
		return nil, ErrNotMember
	}
	for _, role := range roles {
		if mem.Role == role {
			return mem, nil
		}
	}
	return nil, ErrInsufficientRole
}

// RequireOwnerExists verifies that the tenant has at least one owner
// besides the excluded user. Used before demoting or removing an owner.
func RequireOwnerExists(ctx context.Context, memberships Memberships, tenantID, excludeUserID string) error {
	count, err := memberships.CountOtherOwners(ctx, tenantID, excludeUserID)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrLastOwner
	}
	return nil
}
