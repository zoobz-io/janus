package transformers

import (
	"context"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// LicenseToResponse transforms a License model to an admin API response,
// resolving the application's label.
func LicenseToResponse(ctx context.Context, l *models.License, labels contracts.ApplicationLabels) (wire.LicenseResponse, error) {
	names, err := labels.ResolveNames(ctx, []string{l.ApplicationID})
	if err != nil {
		return wire.LicenseResponse{}, err
	}
	return licenseToResponse(l, names), nil
}

// LicensesToResponse transforms a slice of License models to responses,
// resolving all application labels in a single batch.
func LicensesToResponse(ctx context.Context, licenses []*models.License, labels contracts.ApplicationLabels) ([]wire.LicenseResponse, error) {
	ids := make([]string, len(licenses))
	for i, l := range licenses {
		ids[i] = l.ApplicationID
	}
	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]wire.LicenseResponse, len(licenses))
	for i, l := range licenses {
		result[i] = licenseToResponse(l, names)
	}
	return result, nil
}

func licenseToResponse(l *models.License, names map[string]string) wire.LicenseResponse {
	return wire.LicenseResponse{
		ID:          l.ID,
		TenantID:    l.TenantID,
		Application: names[l.ApplicationID],
		CreatedAt:   l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
