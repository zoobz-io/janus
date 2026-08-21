package wire

import "github.com/zoobz-io/check"

// AuthorizeLicenseRequest is the request body for authorizing a tenant on an application.
type AuthorizeLicenseRequest struct {
	TenantID string `json:"tenant_id" description:"Tenant ID to authorize" example:"660e8400-e29b-41d4-a716-446655440000"`
}

// Validate checks the request body.
func (r AuthorizeLicenseRequest) Validate() error {
	return check.Str(r.TenantID, "tenant_id").Required().V().Err()
}

// LicenseResponse is the admin API response for a license.
type LicenseResponse struct {
	ID          string `json:"id" description:"License ID" example:"880e8400-e29b-41d4-a716-446655440000"`
	TenantID    string `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	Application string `json:"application" description:"Application (label)" example:"nexus"`
	CreatedAt   string `json:"created_at" description:"When the license was granted" example:"2026-05-01T12:00:00Z"`
}

// LicenseListResponse is the admin API response for a list of licenses.
type LicenseListResponse struct {
	Licenses []LicenseResponse `json:"licenses" description:"Licenses"`
}

// Clone returns a deep copy of the response.
func (r LicenseListResponse) Clone() LicenseListResponse {
	c := r
	if r.Licenses != nil {
		c.Licenses = make([]LicenseResponse, len(r.Licenses))
		copy(c.Licenses, r.Licenses)
	}
	return c
}
