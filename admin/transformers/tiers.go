package transformers

import (
	"context"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// TierToResponse transforms a Tier model to an admin API response, resolving the
// owning application's label.
func TierToResponse(ctx context.Context, t *models.Tier, labels contracts.ApplicationLabels) (wire.TierResponse, error) {
	names, err := labels.ResolveNames(ctx, []string{t.ApplicationID})
	if err != nil {
		return wire.TierResponse{}, err
	}
	return tierToResponse(t, names), nil
}

// TiersToResponse transforms a slice of Tier models to responses, resolving all
// application labels in a single batch.
func TiersToResponse(ctx context.Context, tiers []*models.Tier, labels contracts.ApplicationLabels) ([]wire.TierResponse, error) {
	ids := make([]string, len(tiers))
	for i, t := range tiers {
		ids[i] = t.ApplicationID
	}
	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]wire.TierResponse, len(tiers))
	for i, t := range tiers {
		result[i] = tierToResponse(t, names)
	}
	return result, nil
}

func tierToResponse(t *models.Tier, names map[string]string) wire.TierResponse {
	return wire.TierResponse{
		ID:          t.ID,
		Application: names[t.ApplicationID],
		Slug:        t.Slug,
		Name:        t.Name,
		Rank:        t.Rank,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
