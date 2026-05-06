package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Memberships defines the contract for membership operations on the public API surface.
type Memberships interface {
	// ListByUser retrieves all memberships for a user.
	ListByUser(ctx context.Context, userID string) ([]*models.Membership, error)
	// ListByTenant retrieves all memberships for a tenant.
	ListByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Membership], error)
	// GetByUserAndTenant retrieves a membership by user and tenant.
	GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*models.Membership, error)
	// Create creates a new membership.
	Create(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error)
	// Delete removes a membership by ID.
	Delete(ctx context.Context, id string) error
}
