package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// TenantApplications defines the contract for tenant-application authorization operations.
type TenantApplications interface {
	// ListByTenant retrieves all application authorizations for a tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]*models.TenantApplication, error)
	// Authorize grants a tenant access to an application.
	Authorize(ctx context.Context, tenantID, applicationID string) (*models.TenantApplication, error)
	// Revoke removes a tenant's access to an application.
	Revoke(ctx context.Context, tenantID, applicationID string) error
}
