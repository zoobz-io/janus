package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// TenantToResponse transforms a Tenant model to an admin API response.
func TenantToResponse(t *models.Tenant) wire.TenantResponse {
	return wire.TenantResponse{
		ID:     t.ID,
		Name:   t.Name,
		Slug:   t.Slug,
		Status: t.Status,
	}
}

// TenantsToResponse transforms a slice of Tenant models to responses.
func TenantsToResponse(tenants []*models.Tenant) []wire.TenantResponse {
	result := make([]wire.TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = TenantToResponse(t)
	}
	return result
}

// ResolveTenantSearch turns a validated tenant search request into store params,
// applying the contract defaults (page 1, size 25, sort updated_at desc). Tenants
// are a root entity, so there is no label resolution. Returns the params plus the
// resolved page number and size.
func ResolveTenantSearch(body wire.SearchTenantsRequest) (stores.TenantSearchParams, int, int) {
	number, size := defaultSearchPageNumber, defaultSearchPageSize
	if p := body.Page; p != nil {
		if p.Number != nil {
			number = *p.Number
		}
		if p.Size != nil {
			size = *p.Size
		}
	}

	sortField, sortOrder := defaultSearchSortField, defaultSearchSortOrder
	if s := body.Sort; s != nil {
		sortField = s.Field
		sortOrder = sortOrderToSQL(s.Order)
	}

	params := stores.TenantSearchParams{
		Query:    body.Query,
		Statuses: body.Facets["status"],
		Dates:    searchDateBounds(body.Dates),
		Sort:     stores.SearchSort{Field: sortField, Order: sortOrder},
		Page:     stores.SearchPage{Offset: (number - 1) * size, Limit: size},
	}
	return params, number, size
}

// TenantSearchToResponse transforms a tenant search result into the response,
// computing total_pages as ceil(total_items / size).
func TenantSearchToResponse(result *stores.TenantSearchResult, number, size int) wire.TenantSearchResponse {
	totalPages := 0
	if size > 0 {
		totalPages = int((result.TotalItems + int64(size) - 1) / int64(size))
	}
	return wire.TenantSearchResponse{
		Tenants: TenantsToResponse(result.Items),
		Page: wire.PageResponse{
			Number:     number,
			Size:       size,
			TotalItems: result.TotalItems,
			TotalPages: totalPages,
		},
		Facets: map[string][]string{"status": result.Statuses},
	}
}
