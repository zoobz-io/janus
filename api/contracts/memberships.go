package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Memberships defines the contract for membership operations on the public API surface.
type Memberships interface {
	// ListByUser retrieves all memberships for a user.
	ListByUser(ctx context.Context, userID string) ([]*models.Membership, error)
	// Create creates a new membership.
	Create(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error)
}
