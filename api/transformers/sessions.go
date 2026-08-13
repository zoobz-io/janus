package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
)

// SessionToResponse transforms a Session model to an API response.
func SessionToResponse(s *models.Session) wire.SessionResponse {
	return wire.SessionResponse{
		ID:        s.ID,
		IssuedBy:  s.IssuedBy,
		UserAgent: s.UserAgent,
		IPAddress: s.IPAddress,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ExpiresAt: s.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}
}

// SessionsToResponse transforms a slice of Session models to responses.
func SessionsToResponse(sessions []*models.Session) []wire.SessionResponse {
	result := make([]wire.SessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = SessionToResponse(s)
	}
	return result
}
