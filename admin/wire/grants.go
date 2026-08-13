package wire

import "github.com/zoobz-io/check"

// CreateGrantRequest is the request body for granting a user access to an application.
type CreateGrantRequest struct {
	UserID   string   `json:"user_id" description:"User ID to grant access to" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID string   `json:"tenant_id" description:"Tenant the access is scoped to" example:"660e8400-e29b-41d4-a716-446655440000"`
	TierID   string   `json:"tier_id,omitempty" description:"Optional tier to place the grant on"`
	Roles    []string `json:"roles" description:"Application-defined roles"`
	Scopes   []string `json:"scopes" description:"Application-defined scopes"`
}

// Validate checks the request body.
func (r CreateGrantRequest) Validate() error {
	return check.All(
		check.Str(r.UserID, "user_id").Required().V(),
		check.Str(r.TenantID, "tenant_id").Required().V(),
	)
}

// UpdateGrantRequest is the request body for updating a grant's roles, scopes and tier.
type UpdateGrantRequest struct {
	TierID string   `json:"tier_id,omitempty" description:"Tier to place the grant on (empty clears it)"`
	Roles  []string `json:"roles" description:"Application-defined roles"`
	Scopes []string `json:"scopes" description:"Application-defined scopes"`
}

// Validate checks the request body.
func (r UpdateGrantRequest) Validate() error {
	return nil
}

// GrantResponse is the admin API response for a grant.
type GrantResponse struct {
	ID            string   `json:"id" description:"Grant ID" example:"990e8400-e29b-41d4-a716-446655440000"`
	UserID        string   `json:"user_id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID      string   `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	ApplicationID string   `json:"application_id" description:"Application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	TierID        string   `json:"tier_id,omitempty" description:"Tier the grant is on, if any"`
	CreatedAt     string   `json:"created_at" description:"When access was granted" example:"2026-05-01T12:00:00Z"`
	Roles         []string `json:"roles" description:"Application-defined roles"`
	Scopes        []string `json:"scopes" description:"Application-defined scopes"`
}

// GrantListResponse is the admin API response for a list of grants.
type GrantListResponse struct {
	Grants []GrantResponse `json:"grants" description:"User grants"`
}

// Clone returns a deep copy of the response.
func (r GrantListResponse) Clone() GrantListResponse {
	c := r
	if r.Grants != nil {
		c.Grants = make([]GrantResponse, len(r.Grants))
		copy(c.Grants, r.Grants)
	}
	return c
}
