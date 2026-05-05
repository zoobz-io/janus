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

// LinkedIdentities provides database access for linked identity records.
type LinkedIdentities struct {
	*sum.Database[models.LinkedIdentity]
}

// NewLinkedIdentities creates a new linked identities store.
func NewLinkedIdentities(db *sqlx.DB, renderer astql.Renderer) *LinkedIdentities {
	return &LinkedIdentities{
		Database: sum.NewDatabase[models.LinkedIdentity](db, "linked_identities", renderer),
	}
}

// ListByUser retrieves all linked identities for a user.
func (s *LinkedIdentities) ListByUser(ctx context.Context, userID string) ([]*models.LinkedIdentity, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		OrderBy("linked_at", "ASC").
		Exec(ctx, map[string]any{"user_id": userID})
}

// GetByProviderSubject looks up a linked identity by provider and external subject.
func (s *LinkedIdentities) GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*models.LinkedIdentity, error) {
	results, err := s.Query().
		Where("provider", "=", "provider").
		Where("external_subject", "=", "external_subject").
		Limit(1).
		Exec(ctx, map[string]any{"provider": provider, "external_subject": externalSubject})
	if err != nil {
		return nil, fmt.Errorf("looking up linked identity: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Link creates a new linked identity record.
func (s *LinkedIdentities) Link(ctx context.Context, userID, provider, externalSubject string) (*models.LinkedIdentity, error) {
	l := &models.LinkedIdentity{
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

// Unlink removes a linked identity by ID, scoped to a user.
func (s *LinkedIdentities) Unlink(ctx context.Context, id, userID string) error {
	results, err := s.Query().
		Where("id", "=", "id").
		Where("user_id", "=", "user_id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id, "user_id": userID})
	if err != nil {
		return fmt.Errorf("finding linked identity: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("linked identity %w", ErrNotFound)
	}
	return s.Delete(ctx, id)
}
