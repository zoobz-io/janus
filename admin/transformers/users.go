package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// UserToResponse transforms a User model to an admin API response.
func UserToResponse(u *models.User) wire.UserResponse {
	lastSeen := ""
	if u.LastSeenAt != nil {
		lastSeen = u.LastSeenAt.Format("2006-01-02T15:04:05Z")
	}
	return wire.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Status:      u.Status,
		LastSeenAt:  lastSeen,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UsersToResponse transforms a slice of User models to responses.
func UsersToResponse(users []*models.User) []wire.UserResponse {
	result := make([]wire.UserResponse, len(users))
	for i, u := range users {
		result[i] = UserToResponse(u)
	}
	return result
}

// ResolveUserSearch turns a validated user search request into store params,
// applying the contract defaults (page 1, size 25, sort updated_at desc). Users
// are a root entity, so there is no label resolution. Returns the params plus
// the resolved page number and size.
func ResolveUserSearch(body wire.SearchUsersRequest) (stores.UserSearchParams, int, int) {
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

	params := stores.UserSearchParams{
		Query:    body.Query,
		Statuses: body.Facets["status"],
		Dates:    searchDateBounds(body.Dates),
		Sort:     stores.SearchSort{Field: sortField, Order: sortOrder},
		Page:     stores.SearchPage{Offset: (number - 1) * size, Limit: size},
	}
	return params, number, size
}

// UserSearchToResponse transforms a user search result into the response,
// computing total_pages as ceil(total_items / size).
func UserSearchToResponse(result *stores.UserSearchResult, number, size int) wire.UserSearchResponse {
	totalPages := 0
	if size > 0 {
		totalPages = int((result.TotalItems + int64(size) - 1) / int64(size))
	}
	return wire.UserSearchResponse{
		Users: UsersToResponse(result.Items),
		Page: wire.PageResponse{
			Number:     number,
			Size:       size,
			TotalItems: result.TotalItems,
			TotalPages: totalPages,
		},
		Facets: map[string][]string{"status": result.Statuses},
	}
}
