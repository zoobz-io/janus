package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/models"
)

// TierToResponse transforms a Tier model to an admin API response.
func TierToResponse(t *models.Tier) wire.TierResponse {
	return wire.TierResponse{
		ID:            t.ID,
		ApplicationID: t.ApplicationID,
		Slug:          t.Slug,
		Name:          t.Name,
		Rank:          t.Rank,
		CreatedAt:     t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// TiersToResponse transforms a slice of Tier models to responses.
func TiersToResponse(tiers []*models.Tier) []wire.TierResponse {
	result := make([]wire.TierResponse, len(tiers))
	for i, t := range tiers {
		result[i] = TierToResponse(t)
	}
	return result
}
