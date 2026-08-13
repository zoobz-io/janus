package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// GrantToResponse transforms a Grant model to an admin API response.
func GrantToResponse(g *models.Grant) wire.GrantResponse {
	tierID := ""
	if g.TierID != nil {
		tierID = *g.TierID
	}
	return wire.GrantResponse{
		ID:            g.ID,
		UserID:        g.UserID,
		TenantID:      g.TenantID,
		ApplicationID: g.ApplicationID,
		TierID:        tierID,
		Roles:         g.Roles,
		Scopes:        g.Scopes,
		CreatedAt:     g.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GrantsToResponse transforms a slice of Grant models to responses.
func GrantsToResponse(grants []*models.Grant) []wire.GrantResponse {
	result := make([]wire.GrantResponse, len(grants))
	for i, g := range grants {
		result[i] = GrantToResponse(g)
	}
	return result
}
