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

// Scopes provides database access for application-defined scopes.
type Scopes struct {
	*sum.Database[models.Scope]
}

// NewScopes creates a new scopes store.
func NewScopes(db *sqlx.DB, renderer astql.Renderer) *Scopes {
	return &Scopes{
		Database: sum.NewDatabase[models.Scope](db, "scopes", renderer),
	}
}

// Define creates a new scope for an application.
func (s *Scopes) Define(ctx context.Context, applicationID, name, description string) (*models.Scope, error) {
	sc := &models.Scope{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		Name:          name,
		Description:   description,
	}
	if err := s.Set(ctx, "", sc); err != nil {
		return nil, fmt.Errorf("defining scope: %w", err)
	}
	return sc, nil
}

// GetByID retrieves a scope by ID. Returns (nil, nil) if absent.
func (s *Scopes) GetByID(ctx context.Context, id string) (*models.Scope, error) {
	results, err := s.Query().
		Where("id", "=", "id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("looking up scope: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// GetByName retrieves a scope by application and name. Returns (nil, nil) if absent.
func (s *Scopes) GetByName(ctx context.Context, applicationID, name string) (*models.Scope, error) {
	results, err := s.Query().
		Where("application_id", "=", "application_id").
		Where("name", "=", "name").
		Limit(1).
		Exec(ctx, map[string]any{"application_id": applicationID, "name": name})
	if err != nil {
		return nil, fmt.Errorf("looking up scope by name: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// ListByApplication retrieves all scopes defined by an application.
func (s *Scopes) ListByApplication(ctx context.Context, applicationID string) ([]*models.Scope, error) {
	return s.Query().
		Where("application_id", "=", "application_id").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"application_id": applicationID})
}

// Delete removes a scope by ID. Features referencing it are cascaded by the database.
func (s *Scopes) Delete(ctx context.Context, id string) error {
	if err := s.Database.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting scope: %w", err)
	}
	return nil
}
