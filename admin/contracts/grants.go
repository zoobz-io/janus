package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Grants defines the admin API's capability boundary over the grants store.
type Grants interface {
	ListByTenantAndApp(ctx context.Context, tenantID, applicationID string) ([]*models.Grant, error)
	Grant(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error)
	UpdateAccess(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error)
	SetTier(ctx context.Context, userID, tenantID, applicationID, tierID string) (*models.Grant, error)
	Revoke(ctx context.Context, userID, tenantID, applicationID string) error
}
