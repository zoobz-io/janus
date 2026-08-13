package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
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

// LicenseToResponse transforms a License model to an API response.
func LicenseToResponse(ta *models.License) wire.LicenseResponse {
	return wire.LicenseResponse{
		ID:            ta.ID,
		TenantID:      ta.TenantID,
		ApplicationID: ta.ApplicationID,
		CreatedAt:     ta.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// LicensesToResponse transforms a slice of License models to responses.
func LicensesToResponse(tas []*models.License) []wire.LicenseResponse {
	result := make([]wire.LicenseResponse, len(tas))
	for i, ta := range tas {
		result[i] = LicenseToResponse(ta)
	}
	return result
}

// GrantToResponse transforms a Grant model to an API response.
func GrantToResponse(ua *models.Grant) wire.GrantResponse {
	return wire.GrantResponse{
		ID:            ua.ID,
		UserID:        ua.UserID,
		TenantID:      ua.TenantID,
		ApplicationID: ua.ApplicationID,
		CreatedAt:     ua.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GrantsToResponse transforms a slice of Grant models to responses.
func GrantsToResponse(uas []*models.Grant) []wire.GrantResponse {
	result := make([]wire.GrantResponse, len(uas))
	for i, ua := range uas {
		result[i] = GrantToResponse(ua)
	}
	return result
}
