package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Licenses defines the admin API's capability boundary over the licenses store.
type Licenses interface {
	ListByApplication(ctx context.Context, applicationID string) ([]*models.License, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*models.License, error)
	Authorize(ctx context.Context, tenantID, applicationID string) (*models.License, error)
	Revoke(ctx context.Context, tenantID, applicationID string) error
}
