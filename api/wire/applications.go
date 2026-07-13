package wire

import "github.com/zoobz-io/check"

// AuthorizeApplicationRequest is the request body for authorizing a tenant to use an application.
type AuthorizeApplicationRequest struct {
	ApplicationID string `json:"application_id" description:"Application ID to authorize" example:"770e8400-e29b-41d4-a716-446655440000"`
}

// Validate checks the request body.
func (r AuthorizeApplicationRequest) Validate() error {
	return check.Str(r.ApplicationID, "application_id").Required().V()
}

// GrantApplicationRequest is the request body for granting a user access to an application.
type GrantApplicationRequest struct {
	ApplicationID string `json:"application_id" description:"Application ID to grant" example:"770e8400-e29b-41d4-a716-446655440000"`
	UserID        string `json:"user_id" description:"User ID to grant access to" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Validate checks the request body.
func (r GrantApplicationRequest) Validate() error {
	return check.All(
		check.Str(r.ApplicationID, "application_id").Required().V(),
		check.Str(r.UserID, "user_id").Required().V(),
	)
}

// ApplicationResponse is the public API response for an application.
type ApplicationResponse struct {
	ID   string `json:"id" description:"Application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" description:"Application name" example:"Nexus"`
	Slug string `json:"slug" description:"Application slug" example:"nexus"`
}

// ApplicationListResponse is the public API response for a list of applications.
type ApplicationListResponse struct {
	Applications []ApplicationResponse `json:"applications" description:"Available applications"`
}

// Clone returns a deep copy of the response.
func (r ApplicationListResponse) Clone() ApplicationListResponse {
	c := r
	if r.Applications != nil {
		c.Applications = make([]ApplicationResponse, len(r.Applications))
		copy(c.Applications, r.Applications)
	}
	return c
}

// LicenseResponse is the public API response for a license authorization.
type LicenseResponse struct {
	ID            string `json:"id" description:"Authorization ID" example:"880e8400-e29b-41d4-a716-446655440000"`
	TenantID      string `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	ApplicationID string `json:"application_id" description:"Application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	CreatedAt     string `json:"created_at" description:"When the authorization was granted" example:"2026-05-01T12:00:00Z"`
}

// LicenseListResponse is the public API response for a list of license authorizations.
type LicenseListResponse struct {
	Authorizations []LicenseResponse `json:"authorizations" description:"Tenant application authorizations"`
}

// Clone returns a deep copy of the response.
func (r LicenseListResponse) Clone() LicenseListResponse {
	c := r
	if r.Authorizations != nil {
		c.Authorizations = make([]LicenseResponse, len(r.Authorizations))
		copy(c.Authorizations, r.Authorizations)
	}
	return c
}

// GrantResponse is the public API response for a grant.
type GrantResponse struct {
	ID            string `json:"id" description:"Grant ID" example:"990e8400-e29b-41d4-a716-446655440000"`
	UserID        string `json:"user_id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID      string `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	ApplicationID string `json:"application_id" description:"Application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	CreatedAt     string `json:"created_at" description:"When access was granted" example:"2026-05-01T12:00:00Z"`
}

// GrantListResponse is the public API response for a list of grants.
type GrantListResponse struct {
	Grants []GrantResponse `json:"grants" description:"User application grants"`
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
