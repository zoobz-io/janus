package wire

import "github.com/zoobz-io/check"

// AddFeatureRequest is the request body for bundling a scope into a tier.
type AddFeatureRequest struct {
	ScopeID string `json:"scope_id" description:"Scope ID to bundle into the tier" example:"aa0e8400-e29b-41d4-a716-446655440000"`
}

// Validate checks the request body.
func (r AddFeatureRequest) Validate() error {
	return check.Str(r.ScopeID, "scope_id").Required().V().Err()
}

// FeatureResponse is the admin API response for a scope bundled into a tier.
type FeatureResponse struct {
	ID        string `json:"id" description:"Feature ID" example:"cc0e8400-e29b-41d4-a716-446655440000"`
	TierID    string `json:"tier_id" description:"Tier ID" example:"bb0e8400-e29b-41d4-a716-446655440000"`
	ScopeID   string `json:"scope_id" description:"Scope ID" example:"aa0e8400-e29b-41d4-a716-446655440000"`
	CreatedAt string `json:"created_at" description:"When the scope was bundled" example:"2026-07-12T12:00:00Z"`
}

// FeatureListResponse is the admin API response for a tier's bundled scopes.
type FeatureListResponse struct {
	Features []FeatureResponse `json:"features" description:"Scopes bundled into the tier"`
}

// Clone returns a deep copy of the response.
func (r FeatureListResponse) Clone() FeatureListResponse {
	c := r
	if r.Features != nil {
		c.Features = make([]FeatureResponse, len(r.Features))
		copy(c.Features, r.Features)
	}
	return c
}
