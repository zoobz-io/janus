package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// AuthorizationToResponse transforms a resolved application authorization to
// an API response.
func AuthorizationToResponse(userID string, app *models.Application, tenants []models.AuthorizedTenant) wire.AuthorizationResponse {
	ats := make([]wire.AuthorizedTenantResponse, len(tenants))
	for i, t := range tenants {
		ats[i] = wire.AuthorizedTenantResponse{
			TenantID:   t.TenantID,
			TenantName: t.TenantName,
			Role:       t.Role,
			Roles:      t.AppRoles,
			Scopes:     t.AppScopes,
		}
	}
	return wire.AuthorizationResponse{
		UserID: userID,
		Application: wire.ApplicationResponse{
			ID:   app.ID,
			Name: app.Name,
			Slug: app.Slug,
		},
		Tenants: ats,
	}
}
