package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Licenses defines the contract for license authorization operations.
type Licenses interface {
	// ListByTenant retrieves all application authorizations for a tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]*models.License, error)
	// Authorize grants a tenant access to an application.
	Authorize(ctx context.Context, tenantID, applicationID string) (*models.License, error)
	// Revoke removes a tenant's access to an application.
	Revoke(ctx context.Context, tenantID, applicationID string) error
}
