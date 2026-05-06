package handlers

import (
	"context"

	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// pathID extracts a string path parameter.
func pathID(params *rocco.Params, name string) string {
	return params.Path[name]
}

// requireRole verifies the caller has at least one of the given roles in the
// specified tenant. Returns the caller's membership or an error.
func requireRole(ctx context.Context, userID, tenantID string, roles ...models.UserRole) (*models.Membership, error) {
	memberships := sum.MustUse[contracts.Memberships](ctx)
	mem, err := memberships.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	if mem == nil {
		return nil, ErrTenantNotFound
	}
	for _, role := range roles {
		if mem.Role == role {
			return mem, nil
		}
	}
	return nil, ErrInsufficientRole
}
