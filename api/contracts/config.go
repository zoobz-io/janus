package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Config defines the contract for runtime configuration access.
type Config interface {
	// GetByKey retrieves a config value by key. Returns (nil, nil) if absent.
	GetByKey(ctx context.Context, key string) (*models.Config, error)
	// Upsert inserts or updates the config value for the given key.
	Upsert(ctx context.Context, key, value string) (*models.Config, error)
}
