package wire

import "github.com/zoobz-io/check"

// CreateTierRequest is the request body for defining a subscription tier.
type CreateTierRequest struct {
	Slug string `json:"slug" description:"URL-safe tier slug" example:"pro"`
	Name string `json:"name" description:"Tier display name" example:"Pro"`
	Rank int    `json:"rank" description:"Ordering rank (ascending)" example:"1"`
}

// Validate checks the request body.
func (r CreateTierRequest) Validate() error {
	return check.All(
		check.Str(r.Slug, "slug").Required().MaxLen(100).V(),
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
	)
}

// TierResponse is the admin API response for a subscription tier.
type TierResponse struct {
	ID            string `json:"id" description:"Tier ID" example:"bb0e8400-e29b-41d4-a716-446655440000"`
	ApplicationID string `json:"application_id" description:"Owning application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	Slug          string `json:"slug" description:"Tier slug" example:"pro"`
	Name          string `json:"name" description:"Tier display name" example:"Pro"`
	CreatedAt     string `json:"created_at" description:"When the tier was defined" example:"2026-07-12T12:00:00Z"`
	Rank          int    `json:"rank" description:"Ordering rank" example:"1"`
}

// TierListResponse is the admin API response for a list of tiers.
type TierListResponse struct {
	Tiers []TierResponse `json:"tiers" description:"Application subscription tiers"`
}

// Clone returns a deep copy of the response.
func (r TierListResponse) Clone() TierListResponse {
	c := r
	if r.Tiers != nil {
		c.Tiers = make([]TierResponse, len(r.Tiers))
		copy(c.Tiers, r.Tiers)
	}
	return c
}
