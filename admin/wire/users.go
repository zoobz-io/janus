package wire

import "github.com/zoobz-io/check"

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Email       string `json:"email" description:"Email address" example:"jane@example.com"`
	DisplayName string `json:"display_name" description:"Display name" example:"Jane Doe"`
}

// Validate checks the request body.
func (r CreateUserRequest) Validate() error {
	return check.All(
		check.Str(r.Email, "email").Required().MaxLen(255).V(),
		check.Str(r.DisplayName, "display_name").Required().MaxLen(255).V(),
	)
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	DisplayName string `json:"display_name" description:"Display name" example:"Jane Doe"`
	Status      string `json:"status" description:"Account status" example:"active"`
}

// Validate checks the request body.
func (r UpdateUserRequest) Validate() error {
	return check.All(
		check.Str(r.DisplayName, "display_name").Required().MaxLen(255).V(),
		check.Str(r.Status, "status").Required().OneOf([]string{"active", "inactive"}).V(),
	)
}

// UserResponse is the admin API response for a user.
type UserResponse struct {
	ID          string `json:"id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email       string `json:"email" description:"Email address" example:"jane@example.com"`
	DisplayName string `json:"display_name" description:"Display name" example:"Jane Doe"`
	Status      string `json:"status" description:"Account status" example:"active"`
	LastSeenAt  string `json:"last_seen_at,omitempty" description:"When the user was last seen" example:"2026-07-12T12:00:00Z"`
	CreatedAt   string `json:"created_at" description:"When the user was created" example:"2026-05-01T12:00:00Z"`
}

// Clone returns a copy of the response.
func (r UserResponse) Clone() UserResponse {
	return r
}

// UserListResponse is the admin API response for a list of users.
type UserListResponse struct {
	Users []UserResponse `json:"users" description:"Users"`
	Total int64          `json:"total" description:"Total user count" example:"128"`
}

// Clone returns a deep copy of the response.
func (r UserListResponse) Clone() UserListResponse {
	c := r
	if r.Users != nil {
		c.Users = make([]UserResponse, len(r.Users))
		copy(c.Users, r.Users)
	}
	return c
}
