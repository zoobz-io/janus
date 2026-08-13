package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

// Features provides database access for the tier↔scope bundle (features).
type Features struct {
	*sum.Database[models.Feature]
}

// NewFeatures creates a new features store.
func NewFeatures(db *sqlx.DB, renderer astql.Renderer) *Features {
	return &Features{
		Database: sum.NewDatabase[models.Feature](db, "features", renderer),
	}
}

// Add bundles a scope into a tier. If the scope is already in the tier the
// existing feature is returned unchanged.
func (s *Features) Add(ctx context.Context, tierID, scopeID string) (*models.Feature, error) {
	existing, err := s.get(ctx, tierID, scopeID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	f := &models.Feature{
		ID:      uuid.New().String(),
		TierID:  tierID,
		ScopeID: scopeID,
	}
	if err := s.Set(ctx, "", f); err != nil {
		return nil, fmt.Errorf("adding feature: %w", err)
	}
	return f, nil
}

// Remove unbundles a scope from a tier. Returns ErrNotFound if the scope is
// not part of the tier.
func (s *Features) Remove(ctx context.Context, tierID, scopeID string) error {
	f, err := s.get(ctx, tierID, scopeID)
	if err != nil {
		return err
	}
	if f == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, f.ID); err != nil {
		return fmt.Errorf("removing feature: %w", err)
	}
	return nil
}

// ListByTier retrieves all features (bundled scopes) for a tier.
func (s *Features) ListByTier(ctx context.Context, tierID string) ([]*models.Feature, error) {
	return s.Query().
		Where("tier_id", "=", "tier_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"tier_id": tierID})
}

// get looks up a single feature by tier and scope. Returns (nil, nil) if absent.
func (s *Features) get(ctx context.Context, tierID, scopeID string) (*models.Feature, error) {
	results, err := s.Query().
		Where("tier_id", "=", "tier_id").
		Where("scope_id", "=", "scope_id").
		Limit(1).
		Exec(ctx, map[string]any{"tier_id": tierID, "scope_id": scopeID})
	if err != nil {
		return nil, fmt.Errorf("looking up feature: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}
