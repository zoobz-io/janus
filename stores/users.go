package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/models"
)

// Users provides database access for users.
type Users struct {
	*sum.Database[models.User]
}

// NewUsers creates a new users store.
func NewUsers(db *sqlx.DB, renderer astql.Renderer) *Users {
	return &Users{
		Database: sum.NewDatabase[models.User](db, "users", renderer),
	}
}

// GetUser retrieves a user by ID.
func (s *Users) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetUserByEmail retrieves a user by email address.
func (s *Users) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.Select().
		Where("email", "=", "email").
		Exec(ctx, map[string]any{"email": email})
}

// ListUsersByTenant retrieves users for a specific tenant.
func (s *Users) ListUsersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.User]{Items: items, Offset: page.Offset}, nil
}

// CreateUser creates a new user within a tenant.
func (s *Users) CreateUser(ctx context.Context, tenantID, email, displayName string) (*models.User, error) {
	now := time.Now()
	u := &models.User{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Email:       email,
		DisplayName: displayName,
		Role:        models.UserRoleViewer,
		Status:      models.UserStatusActive,
		LastSeenAt:  &now,
	}
	if err := s.Set(ctx, "", u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}

// UpdateDisplayName updates a user's display name.
func (s *Users) UpdateDisplayName(ctx context.Context, id, displayName string) (*models.User, error) {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	u.DisplayName = displayName
	if err := s.Set(ctx, id, u); err != nil {
		return nil, fmt.Errorf("updating user display name: %w", err)
	}
	return u, nil
}

// TouchLastSeen updates the user's last_seen_at timestamp.
func (s *Users) TouchLastSeen(ctx context.Context, id string) error {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now()
	u.LastSeenAt = &now
	if err := s.Set(ctx, id, u); err != nil {
		return fmt.Errorf("updating last seen: %w", err)
	}
	return nil
}

// UpsertFromResolve inserts a user or updates last_seen_at on conflict with email + tenant_id.
func (s *Users) UpsertFromResolve(ctx context.Context, tenantID, email, displayName string) (*models.User, bool, error) {
	now := time.Now()
	u := &models.User{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Email:       email,
		DisplayName: displayName,
		Role:        models.UserRoleViewer,
		Status:      models.UserStatusActive,
		LastSeenAt:  &now,
	}
	result, err := s.Insert().
		OnConflict("email", "tenant_id").
		DoUpdate().
		Set("last_seen_at", "last_seen_at").
		Set("display_name", "display_name").
		Set("updated_at", "updated_at").
		Exec(ctx, u)
	if err != nil {
		return nil, false, fmt.Errorf("upserting user from resolve: %w", err)
	}
	created := result.ID == u.ID
	return result, created, nil
}
