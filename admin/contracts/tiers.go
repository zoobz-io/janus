package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Tiers defines the admin API's capability boundary over the tiers store.
type Tiers interface {
	Define(ctx context.Context, applicationID, slug, name string, rank int) (*models.Tier, error)
	GetByID(ctx context.Context, id string) (*models.Tier, error)
	ListByApplication(ctx context.Context, applicationID string) ([]*models.Tier, error)
	Delete(ctx context.Context, id string) error
}
