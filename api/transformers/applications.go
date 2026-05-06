package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// ApplicationToResponse transforms an Application model to an API response.
func ApplicationToResponse(a *models.Application) wire.ApplicationResponse {
	return wire.ApplicationResponse{
		ID:   a.ID,
		Name: a.Name,
		Slug: a.Slug,
	}
}

// ApplicationsToResponse transforms a slice of Application models to responses.
func ApplicationsToResponse(apps []*models.Application) []wire.ApplicationResponse {
	result := make([]wire.ApplicationResponse, len(apps))
	for i, a := range apps {
		result[i] = ApplicationToResponse(a)
	}
	return result
}

// TenantApplicationToResponse transforms a TenantApplication model to an API response.
func TenantApplicationToResponse(ta *models.TenantApplication) wire.TenantApplicationResponse {
	return wire.TenantApplicationResponse{
		ID:            ta.ID,
		TenantID:      ta.TenantID,
		ApplicationID: ta.ApplicationID,
		CreatedAt:     ta.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// TenantApplicationsToResponse transforms a slice of TenantApplication models to responses.
func TenantApplicationsToResponse(tas []*models.TenantApplication) []wire.TenantApplicationResponse {
	result := make([]wire.TenantApplicationResponse, len(tas))
	for i, ta := range tas {
		result[i] = TenantApplicationToResponse(ta)
	}
	return result
}

// UserApplicationToResponse transforms a UserApplication model to an API response.
func UserApplicationToResponse(ua *models.UserApplication) wire.UserApplicationResponse {
	return wire.UserApplicationResponse{
		ID:            ua.ID,
		UserID:        ua.UserID,
		TenantID:      ua.TenantID,
		ApplicationID: ua.ApplicationID,
		CreatedAt:     ua.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UserApplicationsToResponse transforms a slice of UserApplication models to responses.
func UserApplicationsToResponse(uas []*models.UserApplication) []wire.UserApplicationResponse {
	result := make([]wire.UserApplicationResponse, len(uas))
	for i, ua := range uas {
		result[i] = UserApplicationToResponse(ua)
	}
	return result
}
