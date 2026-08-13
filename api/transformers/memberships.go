package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
)

// MemberToResponse transforms a Membership model to an API response.
func MemberToResponse(m *models.Membership) wire.MemberResponse {
	return wire.MemberResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		TenantID:  m.TenantID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// MembersToResponse transforms a slice of Membership models to responses.
func MembersToResponse(memberships []*models.Membership) []wire.MemberResponse {
	result := make([]wire.MemberResponse, len(memberships))
	for i, m := range memberships {
		result[i] = MemberToResponse(m)
	}
	return result
}
