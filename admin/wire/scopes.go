package wire

import "github.com/zoobz-io/check"

// CreateScopeRequest is the request body for defining an application scope.
type CreateScopeRequest struct {
	Name        string `json:"name" description:"Scope name" example:"projects:read"`
	Description string `json:"description" description:"Human-readable description" example:"Read projects"`
}

// Validate checks the request body.
func (r CreateScopeRequest) Validate() error {
	return check.Str(r.Name, "name").Required().MaxLen(255).V()
}

// ScopeResponse is the admin API response for a scope.
type ScopeResponse struct {
	ID            string `json:"id" description:"Scope ID" example:"aa0e8400-e29b-41d4-a716-446655440000"`
	ApplicationID string `json:"application_id" description:"Owning application ID" example:"770e8400-e29b-41d4-a716-446655440000"`
	Name          string `json:"name" description:"Scope name" example:"projects:read"`
	Description   string `json:"description" description:"Human-readable description" example:"Read projects"`
	CreatedAt     string `json:"created_at" description:"When the scope was defined" example:"2026-07-12T12:00:00Z"`
}

// ScopeListResponse is the admin API response for a list of scopes.
type ScopeListResponse struct {
	Scopes []ScopeResponse `json:"scopes" description:"Application scopes"`
}

// Clone returns a deep copy of the response.
func (r ScopeListResponse) Clone() ScopeListResponse {
	c := r
	if r.Scopes != nil {
		c.Scopes = make([]ScopeResponse, len(r.Scopes))
		copy(c.Scopes, r.Scopes)
	}
	return c
}
