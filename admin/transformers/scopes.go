package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/models"
)

// ScopeToResponse transforms a Scope model to an admin API response.
func ScopeToResponse(s *models.Scope) wire.ScopeResponse {
	return wire.ScopeResponse{
		ID:            s.ID,
		ApplicationID: s.ApplicationID,
		Name:          s.Name,
		Description:   s.Description,
		CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ScopesToResponse transforms a slice of Scope models to responses.
func ScopesToResponse(scopes []*models.Scope) []wire.ScopeResponse {
	result := make([]wire.ScopeResponse, len(scopes))
	for i, s := range scopes {
		result[i] = ScopeToResponse(s)
	}
	return result
}
