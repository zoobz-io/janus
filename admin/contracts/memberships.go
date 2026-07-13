package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Memberships defines the admin API's capability boundary over the memberships store.
type Memberships interface {
	ListByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Membership], error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*models.Membership, error)
	Create(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error)
	UpdateRole(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error)
	Remove(ctx context.Context, userID, tenantID string) error
}
