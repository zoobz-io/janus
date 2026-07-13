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

// Licenses provides database access for license authorizations.
type Licenses struct {
	*sum.Database[models.License]
}

// NewLicenses creates a new licenses store.
func NewLicenses(db *sqlx.DB, renderer astql.Renderer) *Licenses {
	return &Licenses{
		Database: sum.NewDatabase[models.License](db, "licenses", renderer),
	}
}

// ListByTenant retrieves all application authorizations for a tenant.
func (s *Licenses) ListByTenant(ctx context.Context, tenantID string) ([]*models.License, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
}

// ListByApplication retrieves all tenant authorizations for an application (admin API).
func (s *Licenses) ListByApplication(ctx context.Context, applicationID string) ([]*models.License, error) {
	return s.Query().
		Where("application_id", "=", "application_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"application_id": applicationID})
}

// GetByTenantAndApp retrieves a specific license authorization.
func (s *Licenses) GetByTenantAndApp(ctx context.Context, tenantID, applicationID string) (*models.License, error) {
	results, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		Limit(1).
		Exec(ctx, map[string]any{"tenant_id": tenantID, "application_id": applicationID})
	if err != nil {
		return nil, fmt.Errorf("looking up license: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Authorize grants a tenant access to an application.
func (s *Licenses) Authorize(ctx context.Context, tenantID, applicationID string) (*models.License, error) {
	ta := &models.License{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		ApplicationID: applicationID,
	}
	if err := s.Set(ctx, "", ta); err != nil {
		return nil, fmt.Errorf("authorizing license: %w", err)
	}
	return ta, nil
}

// Revoke removes a tenant's access to an application.
func (s *Licenses) Revoke(ctx context.Context, tenantID, applicationID string) error {
	ta, err := s.GetByTenantAndApp(ctx, tenantID, applicationID)
	if err != nil {
		return err
	}
	if ta == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, ta.ID); err != nil {
		return fmt.Errorf("revoking license: %w", err)
	}
	return nil
}
