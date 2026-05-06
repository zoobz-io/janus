package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// UserApplications defines the contract for user-application delegation operations.
type UserApplications interface {
	// ListByUser retrieves all application grants for a user within a tenant.
	ListByUser(ctx context.Context, userID, tenantID string) ([]*models.UserApplication, error)
	// Grant delegates access to an application for a user within a tenant.
	Grant(ctx context.Context, userID, tenantID, applicationID string) (*models.UserApplication, error)
	// Revoke removes a user's access to an application within a tenant.
	Revoke(ctx context.Context, userID, tenantID, applicationID string) error
}
