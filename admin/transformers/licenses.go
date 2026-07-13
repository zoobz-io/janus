package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/models"
)

// LicenseToResponse transforms a License model to an admin API response.
func LicenseToResponse(l *models.License) wire.LicenseResponse {
	return wire.LicenseResponse{
		ID:            l.ID,
		TenantID:      l.TenantID,
		ApplicationID: l.ApplicationID,
		CreatedAt:     l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// LicensesToResponse transforms a slice of License models to responses.
func LicensesToResponse(licenses []*models.License) []wire.LicenseResponse {
	result := make([]wire.LicenseResponse, len(licenses))
	for i, l := range licenses {
		result[i] = LicenseToResponse(l)
	}
	return result
}
