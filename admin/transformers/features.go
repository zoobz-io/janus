package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// FeatureToResponse transforms a Feature model to an admin API response.
func FeatureToResponse(f *models.Feature) wire.FeatureResponse {
	return wire.FeatureResponse{
		ID:        f.ID,
		TierID:    f.TierID,
		ScopeID:   f.ScopeID,
		CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// FeaturesToResponse transforms a slice of Feature models to responses.
func FeaturesToResponse(features []*models.Feature) []wire.FeatureResponse {
	result := make([]wire.FeatureResponse, len(features))
	for i, f := range features {
		result[i] = FeatureToResponse(f)
	}
	return result
}
