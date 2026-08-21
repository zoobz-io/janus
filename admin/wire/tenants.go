package wire

import "github.com/zoobz-io/check"

// CreateTenantRequest is the request body for creating a tenant.
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

// UpdateTenantRequest is the request body for updating a tenant.
type UpdateTenantRequest struct {
	Name   string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Status string `json:"status" description:"Tenant status" example:"active"`
}

// Validate checks the request body.
func (r UpdateTenantRequest) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Status, "status").Required().OneOf([]string{"active", "suspended"}).V(),
	).Err()
}

// TenantResponse is the admin API response for a tenant.
type TenantResponse struct {
	ID     string `json:"id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	Name   string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug   string `json:"slug" description:"Tenant slug" example:"acme-corp"`
	Status string `json:"status" description:"Tenant status" example:"active"`
}

// Clone returns a copy of the response.
func (r TenantResponse) Clone() TenantResponse {
	return r
}

// TenantListResponse is the admin API response for a list of tenants.
type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants" description:"Tenants"`
	Total   int64            `json:"total" description:"Total tenant count" example:"42"`
}

// Clone returns a deep copy of the response.
func (r TenantListResponse) Clone() TenantListResponse {
	c := r
	if r.Tenants != nil {
		c.Tenants = make([]TenantResponse, len(r.Tenants))
		copy(c.Tenants, r.Tenants)
	}
	return c
}
