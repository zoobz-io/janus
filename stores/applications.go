package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/models"
)

// Applications provides database access for applications.
type Applications struct {
	*sum.Database[models.Application]
}

// NewApplications creates a new applications store.
func NewApplications(db *sqlx.DB, renderer astql.Renderer) *Applications {
	return &Applications{
		Database: sum.NewDatabase[models.Application](db, "applications", renderer),
	}
}

// GetApplication retrieves an application by ID.
func (s *Applications) GetApplication(ctx context.Context, id string) (*models.Application, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetBySlug retrieves an application by slug.
func (s *Applications) GetBySlug(ctx context.Context, slug string) (*models.Application, error) {
	return s.Select().
		Where("slug", "=", "slug").
		Exec(ctx, map[string]any{"slug": slug})
}

// ListApplications retrieves all active applications.
func (s *Applications) ListApplications(ctx context.Context) ([]*models.Application, error) {
	return s.Query().
		Where("status", "=", "status").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"status": models.ApplicationStatusActive})
}

// ListAll retrieves every application regardless of status (admin API).
func (s *Applications) ListAll(ctx context.Context) ([]*models.Application, error) {
	return s.Query().
		OrderBy("name", "ASC").
		Exec(ctx, nil)
}

// Update updates an application's name and status (admin API).
func (s *Applications) Update(ctx context.Context, id, name string, status models.ApplicationStatus) (*models.Application, error) {
	a, err := s.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Name = name
	a.Status = status
	if err := s.Set(ctx, id, a); err != nil {
		return nil, fmt.Errorf("updating application: %w", err)
	}
	return a, nil
}

// CreateApplication creates a new application.
func (s *Applications) CreateApplication(ctx context.Context, name, slug string) (*models.Application, error) {
	a := &models.Application{
		ID:     uuid.New().String(),
		Name:   name,
		Slug:   slug,
		Status: models.ApplicationStatusActive,
	}
	if err := s.Set(ctx, "", a); err != nil {
		return nil, fmt.Errorf("creating application: %w", err)
	}
	return a, nil
}
