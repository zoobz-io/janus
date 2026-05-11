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

// UserApplications provides database access for user-application grants.
type UserApplications struct {
	*sum.Database[models.UserApplication]
}

// NewUserApplications creates a new user applications store.
func NewUserApplications(db *sqlx.DB, renderer astql.Renderer) *UserApplications {
	return &UserApplications{
		Database: sum.NewDatabase[models.UserApplication](db, "user_applications", renderer),
	}
}

// ListByTenantAndApp retrieves all user grants for an application within a tenant.
func (s *UserApplications) ListByTenantAndApp(ctx context.Context, tenantID, applicationID string) ([]*models.UserApplication, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID, "application_id": applicationID})
}

// ListByUser retrieves all application grants for a user within a tenant.
func (s *UserApplications) ListByUser(ctx context.Context, userID, tenantID string) ([]*models.UserApplication, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"user_id": userID, "tenant_id": tenantID})
}

// GetByUserAndApp retrieves a specific user-application grant.
func (s *UserApplications) GetByUserAndApp(ctx context.Context, userID, tenantID, applicationID string) (*models.UserApplication, error) {
	results, err := s.Query().
		Where("user_id", "=", "user_id").
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		Limit(1).
		Exec(ctx, map[string]any{"user_id": userID, "tenant_id": tenantID, "application_id": applicationID})
	if err != nil {
		return nil, fmt.Errorf("looking up user application: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Grant delegates access to an application for a user within a tenant
// with the given application-specific roles and scopes.
func (s *UserApplications) Grant(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.UserApplication, error) {
	if roles == nil {
		roles = []string{}
	}
	if scopes == nil {
		scopes = []string{}
	}
	ua := &models.UserApplication{
		ID:            uuid.New().String(),
		UserID:        userID,
		TenantID:      tenantID,
		ApplicationID: applicationID,
		Roles:         roles,
		Scopes:        scopes,
	}
	if err := s.Set(ctx, "", ua); err != nil {
		return nil, fmt.Errorf("granting user application: %w", err)
	}
	return ua, nil
}

// UpdateAccess updates the roles and scopes on an existing user-application grant.
func (s *UserApplications) UpdateAccess(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.UserApplication, error) {
	ua, err := s.GetByUserAndApp(ctx, userID, tenantID, applicationID)
	if err != nil {
		return nil, err
	}
	if ua == nil {
		return nil, ErrNotFound
	}
	ua.Roles = roles
	ua.Scopes = scopes
	if err := s.Set(ctx, ua.ID, ua); err != nil {
		return nil, fmt.Errorf("updating user application access: %w", err)
	}
	return ua, nil
}

// Revoke removes a user's access to an application within a tenant.
func (s *UserApplications) Revoke(ctx context.Context, userID, tenantID, applicationID string) error {
	ua, err := s.GetByUserAndApp(ctx, userID, tenantID, applicationID)
	if err != nil {
		return err
	}
	if ua == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, ua.ID); err != nil {
		return fmt.Errorf("revoking user application: %w", err)
	}
	return nil
}
