package wire

import "github.com/zoobz-io/check"

// AddMemberRequest is the request body for adding a member to a tenant.
type AddMemberRequest struct {
	UserID string `json:"user_id" description:"User ID to add" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role" description:"Role to assign" example:"viewer"`
}

// Validate checks the request body.
func (r AddMemberRequest) Validate() error {
	return check.All(
		check.Str(r.UserID, "user_id").Required().V(),
		check.Str(r.Role, "role").Required().OneOf([]string{"viewer", "editor", "admin", "owner"}).V(),
	)
}

// UpdateMemberRoleRequest is the request body for updating a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" description:"New role" example:"admin"`
}

// Validate checks the request body.
func (r UpdateMemberRoleRequest) Validate() error {
	return check.Str(r.Role, "role").Required().OneOf([]string{"viewer", "editor", "admin", "owner"}).V()
}

// MemberResponse is the public API response for a tenant member.
type MemberResponse struct {
	ID        string `json:"id" description:"Membership ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string `json:"user_id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID  string `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	Role      string `json:"role" description:"Role" example:"admin"`
	CreatedAt string `json:"created_at" description:"When the membership was created" example:"2026-05-01T12:00:00Z"`
}

// MemberListResponse is the public API response for a list of members.
type MemberListResponse struct {
	Members []MemberResponse `json:"members" description:"Tenant members"`
}

// Clone returns a deep copy of the response.
func (r MemberListResponse) Clone() MemberListResponse {
	c := r
	if r.Members != nil {
		c.Members = make([]MemberResponse, len(r.Members))
		copy(c.Members, r.Members)
	}
	return c
}
