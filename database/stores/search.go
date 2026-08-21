package stores

import (
	"time"

	"github.com/zoobz-io/janus/database/models"
)

// Search query vocabulary. These are not persistence entities — they are the
// resolved, already-defaulted inputs and outputs of the store's Search methods.
// They live with the store (not database/models, and not an API-surface package)
// because the stores are shared across the admin and public APIs.

// SortOrder is a normalized SQL sort direction ("ASC" or "DESC").
type SortOrder = string

// Sort order values.
const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// SearchSort is a resolved sort specification: Field is a store column name and
// Order is "ASC" or "DESC". The handler resolves request input and defaults into
// this before calling the store.
type SearchSort struct {
	Field string
	Order SortOrder
}

// DateBound is an inclusive [From, To] range over a timestamp column. Either end
// is optional; a nil pointer leaves that bound unconstrained.
type DateBound struct {
	From *time.Time
	To   *time.Time
}

// SearchPage holds resolved, clamped pagination. Offset and Limit are already
// defaulted and bounded by the handler before the store runs.
type SearchPage struct {
	Offset int
	Limit  int
}

// ApplicationSearchParams is the resolved input to Applications.Search. All
// filtering is optional; the zero value selects everything. The store receives
// already-defaulted values and never applies request-layer policy.
type ApplicationSearchParams struct {
	// Dates maps a timestamp column ("created_at"/"updated_at") to its inclusive
	// bounds. Absent keys are unconstrained.
	Dates map[string]DateBound
	// Sort is the resolved ordering (defaulted to updated_at DESC by the handler).
	Sort SearchSort
	// Query is the case-insensitive infix text match applied across name/slug.
	Query string
	// Statuses is the OR-set for the status facet. Empty means no status filter.
	Statuses []string
	// Page is the resolved, clamped pagination window.
	Page SearchPage
}

// ApplicationSearchResult carries a page of applications plus the metadata the
// search response needs: TotalItems for page math and Statuses for the distinct
// facet values present in the filtered set.
type ApplicationSearchResult struct {
	Items      []*models.Application
	Statuses   []string
	TotalItems int64
}
