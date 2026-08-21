// Package transformers provides helper functions for the admin API surface:
// mapping between database types and wire types, and resolving requests into
// store query params.
package transformers

import (
	"strings"
	"time"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// Search contract defaults, applied when the request omits the field.
const (
	defaultSearchPageNumber = 1
	defaultSearchPageSize   = 25
	defaultSearchSortField  = "updated_at"
	defaultSearchSortOrder  = stores.SortDesc
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
func ApplicationSearchToResponse(result *stores.ApplicationSearchResult, number, size int) wire.ApplicationSearchResponse {
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

// ResolveApplicationSearch turns a validated search request into store params,
// applying every contract default in one place: page 1, size 25, sort
// updated_at desc. It returns the params plus the resolved page number and size
// so the response transformer can build the page metadata. The request is
// assumed valid (rocco ran Validate first), so number >= 1 and size in [1,100].
func ResolveApplicationSearch(body wire.SearchApplicationsRequest) (stores.ApplicationSearchParams, int, int) {
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

	params := stores.ApplicationSearchParams{
		Query:    body.Query,
		Statuses: body.Facets["status"],
		Dates:    searchDateBounds(body.Dates),
		Sort:     stores.SearchSort{Field: sortField, Order: sortOrder},
		Page:     stores.SearchPage{Offset: (number - 1) * size, Limit: size},
	}
	return params, number, size
}

// sortOrderToSQL maps a request sort order ("asc"/"desc", case-insensitive) to
// the SQL direction. Unrecognized values never reach here — Validate rejects
// them — so anything non-asc defaults to descending.
func sortOrderToSQL(order string) stores.SortOrder {
	if strings.EqualFold(order, "asc") {
		return stores.SortAsc
	}
	return stores.SortDesc
}

// searchDateBounds converts wire date ranges into store date bounds.
func searchDateBounds(in map[string]wire.DateRange) map[string]stores.DateBound {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]stores.DateBound, len(in))
	for field, dr := range in {
		out[field] = stores.DateBound{From: dr.From, To: dr.To}
	}
	return out
}
