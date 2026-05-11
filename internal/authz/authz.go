package authz

import (
	"context"

	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/models"
)

// RequireRole verifies the user has one of the given roles in the tenant.
// Returns the membership on success, or ErrNotMember / ErrInsufficientRole.
func RequireRole(ctx context.Context, memberships contracts.Memberships, userID, tenantID string, roles ...models.UserRole) (*models.Membership, error) {
	mem, err := memberships.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, ErrNotMember
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
func RequireOwnerExists(ctx context.Context, memberships contracts.Memberships, tenantID, excludeUserID string) error {
	result, err := memberships.ListByTenant(ctx, tenantID, models.OffsetPage{Limit: models.MaxPageSize})
	if err != nil {
		return err
	}
	for _, mem := range result.Items {
		if mem.Role == models.UserRoleOwner && mem.UserID != excludeUserID {
			return nil
		}
	}
	return ErrLastOwner
}
