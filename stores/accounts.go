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

// Accounts provides database access for account records.
type Accounts struct {
	*sum.Database[models.Account]
}

// NewAccounts creates a new accounts store.
func NewAccounts(db *sqlx.DB, renderer astql.Renderer) *Accounts {
	return &Accounts{
		Database: sum.NewDatabase[models.Account](db, "accounts", renderer),
	}
}

// ListByUser retrieves all accounts for a user.
func (s *Accounts) ListByUser(ctx context.Context, userID string) ([]*models.Account, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		OrderBy("linked_at", "ASC").
		Exec(ctx, map[string]any{"user_id": userID})
}

// GetByProviderSubject looks up a account by provider and external subject.
func (s *Accounts) GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*models.Account, error) {
	results, err := s.Query().
		Where("provider", "=", "provider").
		Where("external_subject", "=", "external_subject").
		Limit(1).
		Exec(ctx, map[string]any{"provider": provider, "external_subject": externalSubject})
	if err != nil {
		return nil, fmt.Errorf("looking up account: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Link creates a new account record.
func (s *Accounts) Link(ctx context.Context, userID, provider, externalSubject string) (*models.Account, error) {
	l := &models.Account{
		ID:              uuid.New().String(),
		UserID:          userID,
		Provider:        provider,
		ExternalSubject: externalSubject,
	}
	if err := s.Set(ctx, "", l); err != nil {
		return nil, fmt.Errorf("linking identity: %w", err)
	}
	return l, nil
}

// Unlink removes a account by ID, scoped to a user.
func (s *Accounts) Unlink(ctx context.Context, id, userID string) error {
	results, err := s.Query().
		Where("id", "=", "id").
		Where("user_id", "=", "user_id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id, "user_id": userID})
	if err != nil {
		return fmt.Errorf("finding account: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("account %w", ErrNotFound)
	}
	return s.Delete(ctx, id)
}
