package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// LinkedIdentityToResponse transforms a LinkedIdentity model to an API response.
func LinkedIdentityToResponse(l *models.LinkedIdentity) wire.LinkedIdentityResponse {
	return wire.LinkedIdentityResponse{
		ID:              l.ID,
		Provider:        l.Provider,
		ExternalSubject: l.ExternalSubject,
		LinkedAt:        l.LinkedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// LinkedIdentitiesToResponse transforms a slice of LinkedIdentity models to responses.
func LinkedIdentitiesToResponse(identities []*models.LinkedIdentity) []wire.LinkedIdentityResponse {
	result := make([]wire.LinkedIdentityResponse, len(identities))
	for i, l := range identities {
		result[i] = LinkedIdentityToResponse(l)
	}
	return result
}
