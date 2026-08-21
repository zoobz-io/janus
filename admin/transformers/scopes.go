package transformers

import (
	"context"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// ScopeToResponse transforms a Scope model to an admin API response, resolving
// the owning application's label.
func ScopeToResponse(ctx context.Context, s *models.Scope, labels contracts.ApplicationLabels) (wire.ScopeResponse, error) {
	names, err := labels.ResolveNames(ctx, []string{s.ApplicationID})
	if err != nil {
		return wire.ScopeResponse{}, err
	}
	return scopeToResponse(s, names), nil
}

// ScopesToResponse transforms a slice of Scope models to responses, resolving
// all application labels in a single batch.
func ScopesToResponse(ctx context.Context, scopes []*models.Scope, labels contracts.ApplicationLabels) ([]wire.ScopeResponse, error) {
	ids := make([]string, len(scopes))
	for i, s := range scopes {
		ids[i] = s.ApplicationID
	}
	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]wire.ScopeResponse, len(scopes))
	for i, s := range scopes {
		result[i] = scopeToResponse(s, names)
	}
	return result, nil
}

func scopeToResponse(s *models.Scope, names map[string]string) wire.ScopeResponse {
	return wire.ScopeResponse{
		ID:          s.ID,
		Application: names[s.ApplicationID],
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
