package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

// Config provides database access for the runtime config table.
type Config struct {
	*sum.Database[models.Config]
}

// NewConfig creates a new config store.
func NewConfig(db *sqlx.DB, renderer astql.Renderer) *Config {
	return &Config{
		Database: sum.NewDatabase[models.Config](db, "config", renderer),
	}
}

// GetByKey retrieves a config value by key. Returns (nil, nil) if absent.
func (s *Config) GetByKey(ctx context.Context, key string) (*models.Config, error) {
	results, err := s.Query().
		Where("key", "=", "key").
		Limit(1).
		Exec(ctx, map[string]any{"key": key})
	if err != nil {
		return nil, fmt.Errorf("looking up config %q: %w", key, err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Upsert inserts or updates the config value for the given key.
func (s *Config) Upsert(ctx context.Context, key, value string) (*models.Config, error) {
	existing, err := s.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("checking config %q: %w", key, err)
	}

	c := &models.Config{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	if existing == nil {
		if err := s.Set(ctx, "", c); err != nil {
			return nil, fmt.Errorf("inserting config %q: %w", key, err)
		}
		return c, nil
	}

	if err := s.Set(ctx, key, c); err != nil {
		return nil, fmt.Errorf("updating config %q: %w", key, err)
	}
	return c, nil
}
