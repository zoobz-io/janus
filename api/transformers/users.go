// Package transformers provides pure mapping functions between models and wire types.
package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// UserToResponse transforms a User model and its memberships to an API response.
func UserToResponse(u *models.User, memberships []*models.Membership, tenantNames map[string]string) wire.UserResponse {
	mems := make([]wire.MembershipResponse, len(memberships))
	for i, m := range memberships {
		mems[i] = wire.MembershipResponse{
			TenantID:   m.TenantID,
			TenantName: tenantNames[m.TenantID],
			Role:       m.Role,
		}
	}
	return wire.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Memberships: mems,
	}
}
