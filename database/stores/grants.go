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

// Grants provides database access for grant grants.
type Grants struct {
	*sum.Database[models.Grant]
}

// NewGrants creates a new grants store.
func NewGrants(db *sqlx.DB, renderer astql.Renderer) *Grants {
	return &Grants{
		Database: sum.NewDatabase[models.Grant](db, "grants", renderer),
	}
}

// ListByTenantAndApp retrieves all user grants for an application within a tenant.
func (s *Grants) ListByTenantAndApp(ctx context.Context, tenantID, applicationID string) ([]*models.Grant, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID, "application_id": applicationID})
}

// ListByUser retrieves all application grants for a user within a tenant.
func (s *Grants) ListByUser(ctx context.Context, userID, tenantID string) ([]*models.Grant, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"user_id": userID, "tenant_id": tenantID})
}

// GetByUserAndApp retrieves a specific grant grant.
func (s *Grants) GetByUserAndApp(ctx context.Context, userID, tenantID, applicationID string) (*models.Grant, error) {
	results, err := s.Query().
		Where("user_id", "=", "user_id").
		Where("tenant_id", "=", "tenant_id").
		Where("application_id", "=", "application_id").
		Limit(1).
		Exec(ctx, map[string]any{"user_id": userID, "tenant_id": tenantID, "application_id": applicationID})
	if err != nil {
		return nil, fmt.Errorf("looking up grant: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Grant delegates access to an application for a user within a tenant
// with the given application-specific roles and scopes.
func (s *Grants) Grant(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error) {
	if roles == nil {
		roles = []string{}
	}
	if scopes == nil {
		scopes = []string{}
	}
	ua := &models.Grant{
		ID:            uuid.New().String(),
		UserID:        userID,
		TenantID:      tenantID,
		ApplicationID: applicationID,
		Roles:         roles,
		Scopes:        scopes,
	}
	if err := s.Set(ctx, "", ua); err != nil {
		return nil, fmt.Errorf("granting grant: %w", err)
	}
	return ua, nil
}

// UpdateAccess updates the roles and scopes on an existing grant.
func (s *Grants) UpdateAccess(ctx context.Context, userID, tenantID, applicationID string, roles, scopes []string) (*models.Grant, error) {
	ua, err := s.GetByUserAndApp(ctx, userID, tenantID, applicationID)
	if err != nil {
		return nil, err
	}
	if ua == nil {
		return nil, ErrNotFound
	}
	// Nil normalizes to empty so the TEXT[] NOT NULL columns never receive
	// SQL NULL on update — the insert path in Grant guards the same way.
	if roles == nil {
		roles = []string{}
	}
	if scopes == nil {
		scopes = []string{}
	}
	ua.Roles = roles
	ua.Scopes = scopes
	if err := s.Set(ctx, ua.ID, ua); err != nil {
		return nil, fmt.Errorf("updating grant access: %w", err)
	}
	return ua, nil
}

// SetTier places a grant on a tier, or clears it when tierID is empty. The
// caller is responsible for ensuring the tier belongs to the grant's application.
func (s *Grants) SetTier(ctx context.Context, userID, tenantID, applicationID, tierID string) (*models.Grant, error) {
	ua, err := s.GetByUserAndApp(ctx, userID, tenantID, applicationID)
	if err != nil {
		return nil, err
	}
	if ua == nil {
		return nil, ErrNotFound
	}
	if tierID == "" {
		ua.TierID = nil
	} else {
		ua.TierID = &tierID
	}
	if err := s.Set(ctx, ua.ID, ua); err != nil {
		return nil, fmt.Errorf("setting grant tier: %w", err)
	}
	return ua, nil
}

// Revoke removes a user's access to an application within a tenant.
func (s *Grants) Revoke(ctx context.Context, userID, tenantID, applicationID string) error {
	ua, err := s.GetByUserAndApp(ctx, userID, tenantID, applicationID)
	if err != nil {
		return err
	}
	if ua == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, ua.ID); err != nil {
		return fmt.Errorf("revoking grant: %w", err)
	}
	return nil
}
