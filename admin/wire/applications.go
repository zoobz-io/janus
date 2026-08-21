// Package wire defines request and response types for the admin API surface.
package wire

import "github.com/zoobz-io/check"

// CreateApplicationRequest is the request body for registering an application.
type CreateApplicationRequest struct {
	Name string `json:"name" description:"Application name" example:"Nexus"`
	Slug string `json:"slug" description:"URL-safe slug (matches the service cert CN)" example:"nexus"`
}

// Validate checks the request body.
func (r CreateApplicationRequest) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Slug, "slug").Required().MaxLen(100).V(),
	).Err()
}

// UpdateApplicationRequest is the request body for updating an application.
type UpdateApplicationRequest struct {
	Name   string `json:"name" description:"Application name" example:"Nexus"`
	Status string `json:"status" description:"Application status" example:"active"`
}

// Validate checks the request body.
func (r UpdateApplicationRequest) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Status, "status").Required().OneOf([]string{"active", "inactive"}).V(),
	).Err()
}

// ApplicationResponse is the admin API response for an application.
type ApplicationResponse struct {
	ID        string `json:"id" description:"Application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" description:"Application name" example:"Nexus"`
	Slug      string `json:"slug" description:"Application slug" example:"nexus"`
	Status    string `json:"status" description:"Application status" example:"active"`
	CreatedAt string `json:"created_at" description:"Creation timestamp (RFC 3339)" example:"2026-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" description:"Last update timestamp (RFC 3339)" example:"2026-08-21T00:00:00Z"`
}

// Clone returns a copy of the response.
func (r ApplicationResponse) Clone() ApplicationResponse {
	return r
}

// ApplicationListResponse is the admin API response for a list of applications.
type ApplicationListResponse struct {
	Applications []ApplicationResponse `json:"applications" description:"Applications"`
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
