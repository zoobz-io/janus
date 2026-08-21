package transformers

import (
	"context"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// GrantToResponse transforms a Grant model to an admin API response, resolving
// the application's label.
func GrantToResponse(ctx context.Context, g *models.Grant, labels contracts.ApplicationLabels) (wire.GrantResponse, error) {
	names, err := labels.ResolveNames(ctx, []string{g.ApplicationID})
	if err != nil {
		return wire.GrantResponse{}, err
	}
	return grantToResponse(g, names), nil
}

// GrantsToResponse transforms a slice of Grant models to responses, resolving
// all application labels in a single batch.
func GrantsToResponse(ctx context.Context, grants []*models.Grant, labels contracts.ApplicationLabels) ([]wire.GrantResponse, error) {
	ids := make([]string, len(grants))
	for i, g := range grants {
		ids[i] = g.ApplicationID
	}
	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]wire.GrantResponse, len(grants))
	for i, g := range grants {
		result[i] = grantToResponse(g, names)
	}
	return result, nil
}

func grantToResponse(g *models.Grant, names map[string]string) wire.GrantResponse {
	tierID := ""
	if g.TierID != nil {
		tierID = *g.TierID
	}
	return wire.GrantResponse{
		ID:          g.ID,
		UserID:      g.UserID,
		TenantID:    g.TenantID,
		Application: names[g.ApplicationID],
		TierID:      tierID,
		Roles:       g.Roles,
		Scopes:      g.Scopes,
		CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
