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

// TenantApplications provides database access for tenant-application authorizations.
type TenantApplications struct {
	*sum.Database[models.TenantApplication]
}

// NewTenantApplications creates a new tenant applications store.
func NewTenantApplications(db *sqlx.DB, renderer astql.Renderer) *TenantApplications {
	return &TenantApplications{
		Database: sum.NewDatabase[models.TenantApplication](db, "tenant_applications", renderer),
	}
}

// ListByTenant retrieves all application authorizations for a tenant.
func (s *TenantApplications) ListByTenant(ctx context.Context, tenantID string) ([]*models.TenantApplication, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
}

// GetByTenantAndApp retrieves a specific tenant-application authorization.
func (s *TenantApplications) GetByTenantAndApp(ctx context.Context, tenantID, applicationID string) (*models.TenantApplication, error) {
	results, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		Limit(1).
		Exec(ctx, map[string]any{"tenant_id": tenantID, "application_id": applicationID})
	if err != nil {
		return nil, fmt.Errorf("looking up tenant application: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Authorize grants a tenant access to an application.
func (s *TenantApplications) Authorize(ctx context.Context, tenantID, applicationID string) (*models.TenantApplication, error) {
	ta := &models.TenantApplication{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		ApplicationID: applicationID,
	}
	if err := s.Set(ctx, "", ta); err != nil {
		return nil, fmt.Errorf("authorizing tenant application: %w", err)
	}
	return ta, nil
}

// Revoke removes a tenant's access to an application.
func (s *TenantApplications) Revoke(ctx context.Context, tenantID, applicationID string) error {
	ta, err := s.GetByTenantAndApp(ctx, tenantID, applicationID)
	if err != nil {
		return err
	}
	if ta == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, ta.ID); err != nil {
		return fmt.Errorf("revoking tenant application: %w", err)
	}
	return nil
}
