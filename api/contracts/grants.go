package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Grants defines the contract for grant delegation operations.
type Grants interface {
	// ListByUser retrieves all application grants for a user within a tenant.
	ListByUser(ctx context.Context, userID, tenantID string) ([]*models.Grant, error)
	// GetByUserAndApp retrieves a specific grant grant.
	GetByUserAndApp(ctx context.Context, userID, tenantID, applicationID string) (*models.Grant, error)
	// Grant delegates access to an application for a user within a tenant.
	Grant(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error)
	// UpdateAccess updates the roles and scopes on an existing grant.
	UpdateAccess(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error)
	// Revoke removes a user's access to an application within a tenant.
	Revoke(ctx context.Context, userID, tenantID, applicationID string) error
}
