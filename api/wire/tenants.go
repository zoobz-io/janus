package wire

import "github.com/zoobz-io/check"

// CreateTenantRequest is the request body for creating a new tenant.
type CreateTenantRequest struct {
	Name string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug string `json:"slug" description:"URL-safe tenant slug" example:"acme-corp"`
}

// Validate checks the request body.
func (r CreateTenantRequest) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Slug, "slug").Required().MaxLen(100).V(),
	).Err()
}

// TenantResponse is the public API response for a tenant.
type TenantResponse struct {
	ID   string `json:"id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug string `json:"slug" description:"Tenant slug" example:"acme-corp"`
}

// Clone returns a copy of the response.
func (t TenantResponse) Clone() TenantResponse {
	return t
}
