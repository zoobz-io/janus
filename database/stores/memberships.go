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

// Memberships provides database access for tenant memberships.
type Memberships struct {
	*sum.Database[models.Membership]
}

// NewMemberships creates a new memberships store.
func NewMemberships(db *sqlx.DB, renderer astql.Renderer) *Memberships {
	return &Memberships{
		Database: sum.NewDatabase[models.Membership](db, "memberships", renderer),
	}
}

// ListByUser retrieves all memberships for a user.
func (s *Memberships) ListByUser(ctx context.Context, userID string) ([]*models.Membership, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		OrderBy("created_at", "ASC").
		Exec(ctx, map[string]any{"user_id": userID})
}

// ListByTenant retrieves all memberships for a tenant.
func (s *Memberships) ListByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Membership], error) {
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Membership]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// CountOtherOwners counts owners of the tenant other than excludeUserID.
// Owner protection depends on this being a count over the whole table,
// not a scan of one page.
func (s *Memberships) CountOtherOwners(ctx context.Context, tenantID, excludeUserID string) (int64, error) {
	count, err := s.Count().
		Where("tenant_id", "=", "tenant_id").
		Where("role", "=", "role").
		Where("user_id", "!=", "user_id").
		Exec(ctx, map[string]any{"tenant_id": tenantID, "role": models.UserRoleOwner, "user_id": excludeUserID})
	if err != nil {
		return 0, fmt.Errorf("counting tenant owners: %w", err)
	}
	return int64(count), nil
}

// GetByUserAndTenant retrieves a membership by user and tenant.
func (s *Memberships) GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*models.Membership, error) {
	results, err := s.Query().
		Where("user_id", "=", "user_id").
		Where("tenant_id", "=", "tenant_id").
		Limit(1).
		Exec(ctx, map[string]any{"user_id": userID, "tenant_id": tenantID})
	if err != nil {
		return nil, fmt.Errorf("looking up membership: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Create creates a new membership.
func (s *Memberships) Create(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error) {
	m := &models.Membership{
		ID:       uuid.New().String(),
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
	}
	if err := s.Set(ctx, "", m); err != nil {
		return nil, fmt.Errorf("creating membership: %w", err)
	}
	return m, nil
}

// UpdateRole changes a user's role within a tenant. Returns ErrNotFound if the
// membership does not exist.
func (s *Memberships) UpdateRole(ctx context.Context, userID, tenantID string, role models.UserRole) (*models.Membership, error) {
	m, err := s.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	m.Role = role
	if err := s.Set(ctx, m.ID, m); err != nil {
		return nil, fmt.Errorf("updating membership role: %w", err)
	}
	return m, nil
}

// Remove deletes a user's membership in a tenant. Returns ErrNotFound if the
// membership does not exist.
func (s *Memberships) Remove(ctx context.Context, userID, tenantID string) error {
	m, err := s.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrNotFound
	}
	if err := s.Delete(ctx, m.ID); err != nil {
		return fmt.Errorf("removing membership: %w", err)
	}
	return nil
}
