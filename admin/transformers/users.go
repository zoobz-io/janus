package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// UserToResponse transforms a User model to an admin API response.
func UserToResponse(u *models.User) wire.UserResponse {
	lastSeen := ""
	if u.LastSeenAt != nil {
		lastSeen = u.LastSeenAt.Format("2006-01-02T15:04:05Z")
	}
	return wire.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Status:      u.Status,
		LastSeenAt:  lastSeen,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UsersToResponse transforms a slice of User models to responses.
func UsersToResponse(users []*models.User) []wire.UserResponse {
	result := make([]wire.UserResponse, len(users))
	for i, u := range users {
		result[i] = UserToResponse(u)
	}
	return result
}
