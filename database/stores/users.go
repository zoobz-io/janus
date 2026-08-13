package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
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

// CreateUser creates a new user.
func (s *Users) CreateUser(ctx context.Context, email, displayName string) (*models.User, error) {
	now := time.Now()
	u := &models.User{
		ID:          uuid.New().String(),
		Email:       email,
		DisplayName: displayName,
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

// List retrieves a paginated list of all users (admin API).
func (s *Users) List(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().Exec(ctx, nil)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.User]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// Update updates a user's display name and status (admin API).
func (s *Users) Update(ctx context.Context, id, displayName string, status models.UserStatus) (*models.User, error) {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	u.DisplayName = displayName
	u.Status = status
	if err := s.Set(ctx, id, u); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
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
