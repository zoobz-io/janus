package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Tenants defines the admin API's capability boundary over the tenants store.
type Tenants interface {
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	ListTenants(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Tenant], error)
	CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error)
	UpdateTenant(ctx context.Context, id, name string, status models.TenantStatus) (*models.Tenant, error)
}
