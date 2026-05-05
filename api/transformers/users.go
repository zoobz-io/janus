// Package transformers provides pure mapping functions between models and wire types.
package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// UserToResponse transforms a User model to an API response.
func UserToResponse(u *models.User) wire.UserResponse {
	return wire.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		TenantID:    u.TenantID,
	}
}
