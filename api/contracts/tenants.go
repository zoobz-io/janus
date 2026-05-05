package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Tenants defines the contract for tenant operations on the public API surface.
type Tenants interface {
	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	// CreateTenant creates a new tenant.
	CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error)
}
