// Package transformers provides pure mapping functions between models and admin wire types.
package transformers

import (
	"time"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// ApplicationToResponse transforms an Application model to an admin API response.
func ApplicationToResponse(a *models.Application) wire.ApplicationResponse {
	return wire.ApplicationResponse{
		ID:        a.ID,
		Name:      a.Name,
		Slug:      a.Slug,
		Status:    a.Status,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
}

// ApplicationSearchToResponse transforms a search result into the search
// response, computing total_pages as ceil(total_items / size). number and size
// are the resolved (defaulted) pagination the handler ran the query with.
func ApplicationSearchToResponse(result *models.ApplicationSearchResult, number, size int) wire.ApplicationSearchResponse {
	totalPages := 0
	if size > 0 {
		totalPages = int((result.TotalItems + int64(size) - 1) / int64(size))
	}
	return wire.ApplicationSearchResponse{
		Applications: ApplicationsToResponse(result.Items),
		Page: wire.PageResponse{
			Number:     number,
			Size:       size,
			TotalItems: result.TotalItems,
			TotalPages: totalPages,
		},
		Facets: map[string][]string{"status": result.Statuses},
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
