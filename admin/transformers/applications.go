// Package transformers provides pure mapping functions between models and admin wire types.
package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// ApplicationToResponse transforms an Application model to an admin API response.
func ApplicationToResponse(a *models.Application) wire.ApplicationResponse {
	return wire.ApplicationResponse{
		ID:     a.ID,
		Name:   a.Name,
		Slug:   a.Slug,
		Status: a.Status,
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
