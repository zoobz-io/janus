// Package wire defines request and response types for the public API surface.
package wire

import (
	"context"

	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// UserResponse is the public API response for a user profile.
type UserResponse struct {
	Email       string               `json:"email" description:"Email address" example:"jane@example.com" send.mask:"email"`
	DisplayName string               `json:"display_name" description:"Display name" example:"Jane Doe" send.mask:"name"`
	ID          string               `json:"id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	Memberships []MembershipResponse `json:"memberships" description:"Tenant memberships"`
}

// OnSend applies boundary masking before the response is marshaled.
func (u *UserResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[UserResponse]](ctx)
	masked, err := b.Send(ctx, *u)
	if err != nil {
		return err
	}
	*u = masked
	return nil
}

// Clone returns a deep copy of the response.
func (u UserResponse) Clone() UserResponse {
	c := u
	if u.Memberships != nil {
		c.Memberships = make([]MembershipResponse, len(u.Memberships))
		copy(c.Memberships, u.Memberships)
	}
	return c
}

// MembershipResponse is the public API response for a tenant membership.
type MembershipResponse struct {
	TenantID   string `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	TenantName string `json:"tenant_name" description:"Tenant name" example:"Acme Corp"`
	Role       string `json:"role" description:"Role in this tenant" example:"owner"`
}

// UpdateProfileRequest is the request body for updating a user's profile.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" description:"New display name" example:"Jane Doe"`
}

// Validate checks the request body.
func (r UpdateProfileRequest) Validate() error {
	return check.Str(r.DisplayName, "display_name").Required().MaxLen(255).V().Err()
}
