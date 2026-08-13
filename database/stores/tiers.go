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

// Tiers provides database access for application-defined subscription tiers.
type Tiers struct {
	*sum.Database[models.Tier]
}

// NewTiers creates a new tiers store.
func NewTiers(db *sqlx.DB, renderer astql.Renderer) *Tiers {
	return &Tiers{
		Database: sum.NewDatabase[models.Tier](db, "tiers", renderer),
	}
}

// Define creates a new tier for an application.
func (s *Tiers) Define(ctx context.Context, applicationID, slug, name string, rank int) (*models.Tier, error) {
	t := &models.Tier{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		Slug:          slug,
		Name:          name,
		Rank:          rank,
	}
	if err := s.Set(ctx, "", t); err != nil {
		return nil, fmt.Errorf("defining tier: %w", err)
	}
	return t, nil
}

// GetByID retrieves a tier by ID. Returns (nil, nil) if absent.
func (s *Tiers) GetByID(ctx context.Context, id string) (*models.Tier, error) {
	results, err := s.Query().
		Where("id", "=", "id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("looking up tier: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// GetBySlug retrieves a tier by application and slug. Returns (nil, nil) if absent.
func (s *Tiers) GetBySlug(ctx context.Context, applicationID, slug string) (*models.Tier, error) {
	results, err := s.Query().
		Where("application_id", "=", "application_id").
		Where("slug", "=", "slug").
		Limit(1).
		Exec(ctx, map[string]any{"application_id": applicationID, "slug": slug})
	if err != nil {
		return nil, fmt.Errorf("looking up tier by slug: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// ListByApplication retrieves all tiers defined by an application, ordered by rank.
func (s *Tiers) ListByApplication(ctx context.Context, applicationID string) ([]*models.Tier, error) {
	return s.Query().
		Where("application_id", "=", "application_id").
		OrderBy("rank", "ASC").
		Exec(ctx, map[string]any{"application_id": applicationID})
}

// Delete removes a tier by ID. Features referencing it are cascaded by the database.
func (s *Tiers) Delete(ctx context.Context, id string) error {
	if err := s.Database.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting tier: %w", err)
	}
	return nil
}
