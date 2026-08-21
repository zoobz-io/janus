package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Tenants defines the contract for tenant operations on the public API surface.
type Tenants interface {
	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	// CreateTenant creates a new tenant.
	CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error)
}

// Provisioning defines the transactional multi-store flows the public API
// depends on. Satisfied structurally by the store aggregate — these must not
// be recomposed from single-store calls, which would reintroduce the
// partial-write windows they exist to close.
type Provisioning interface {
	// CreateTenantWithOwner creates a tenant and its owner membership atomically.
	CreateTenantWithOwner(ctx context.Context, name, slug, ownerUserID string) (*models.Tenant, *models.Membership, error)
}
