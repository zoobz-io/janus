package transformers

import (
	"context"
	"sort"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// ScopeToResponse transforms a Scope model to an admin API response, resolving
// the owning application's label.
func ScopeToResponse(ctx context.Context, s *models.Scope, labels contracts.ApplicationLabels) (wire.ScopeResponse, error) {
	names, err := labels.ResolveNames(ctx, []string{s.ApplicationID})
	if err != nil {
		return wire.ScopeResponse{}, err
	}
	return scopeToResponse(s, names), nil
}

// ScopesToResponse transforms a slice of Scope models to responses, resolving
// all application labels in a single batch.
func ScopesToResponse(ctx context.Context, scopes []*models.Scope, labels contracts.ApplicationLabels) ([]wire.ScopeResponse, error) {
	ids := make([]string, len(scopes))
	for i, s := range scopes {
		ids[i] = s.ApplicationID
	}
	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]wire.ScopeResponse, len(scopes))
	for i, s := range scopes {
		result[i] = scopeToResponse(s, names)
	}
	return result, nil
}

// ResolveScopeSearch turns a validated scope search request into store params,
// applying the contract defaults. The application facet is resolved inbound from
// names to ids: a present facet whose names are all unknown yields a non-nil
// empty id set (matches nothing), while an absent facet yields nil (no filter).
// Returns the params plus the resolved page number and size.
func ResolveScopeSearch(ctx context.Context, body wire.SearchScopesRequest, labels contracts.ApplicationLabels) (stores.ScopeSearchParams, int, int, error) {
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

	var appIDs []string
	if names, ok := body.Facets["application"]; ok {
		appIDs = []string{} // non-nil: filter requested, even if nothing resolves
		for _, name := range names {
			id, found, err := labels.ResolveID(ctx, name)
			if err != nil {
				return stores.ScopeSearchParams{}, 0, 0, err
			}
			if found {
				appIDs = append(appIDs, id)
			}
		}
	}

	params := stores.ScopeSearchParams{
		Query:          body.Query,
		ApplicationIDs: appIDs,
		Dates:          searchDateBounds(body.Dates),
		Sort:           stores.SearchSort{Field: sortField, Order: sortOrder},
		Page:           stores.SearchPage{Offset: (number - 1) * size, Limit: size},
	}
	return params, number, size, nil
}

// ScopeSearchToResponse assembles the scope search response. It resolves every
// application label in a single batch covering both the row ids and the facet
// ids, then sorts the facet by name (resolution destroys the store's id order).
func ScopeSearchToResponse(ctx context.Context, result *stores.ScopeSearchResult, number, size int, labels contracts.ApplicationLabels) (wire.ScopeSearchResponse, error) {
	ids := make([]string, 0, len(result.Items)+len(result.ApplicationIDs))
	for _, sc := range result.Items {
		ids = append(ids, sc.ApplicationID)
	}
	ids = append(ids, result.ApplicationIDs...)

	names, err := labels.ResolveNames(ctx, ids)
	if err != nil {
		return wire.ScopeSearchResponse{}, err
	}

	scopes := make([]wire.ScopeResponse, len(result.Items))
	for i, sc := range result.Items {
		scopes[i] = scopeToResponse(sc, names)
	}

	facetNames := make([]string, 0, len(result.ApplicationIDs))
	for _, id := range result.ApplicationIDs {
		if name, ok := names[id]; ok {
			facetNames = append(facetNames, name)
		}
	}
	sort.Strings(facetNames)

	totalPages := 0
	if size > 0 {
		totalPages = int((result.TotalItems + int64(size) - 1) / int64(size))
	}
	return wire.ScopeSearchResponse{
		Scopes: scopes,
		Page: wire.PageResponse{
			Number:     number,
			Size:       size,
			TotalItems: result.TotalItems,
			TotalPages: totalPages,
		},
		Facets: map[string][]string{"application": facetNames},
	}, nil
}

func scopeToResponse(s *models.Scope, names map[string]string) wire.ScopeResponse {
	return wire.ScopeResponse{
		ID:          s.ID,
		Application: names[s.ApplicationID],
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
