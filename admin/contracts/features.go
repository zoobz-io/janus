package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Features defines the admin API's capability boundary over the features store.
type Features interface {
	Add(ctx context.Context, tierID, scopeID string) (*models.Feature, error)
	Remove(ctx context.Context, tierID, scopeID string) error
	ListByTier(ctx context.Context, tierID string) ([]*models.Feature, error)
}
