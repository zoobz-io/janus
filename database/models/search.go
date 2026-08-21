package models

import "time"

// SortOrder is a normalized SQL sort direction ("ASC" or "DESC").
type SortOrder = string

// Sort order values.
const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// SearchSort is a resolved sort specification for a search query. Field is a
// store column name; Order is "ASC" or "DESC". The handler resolves request
// input (and defaults) into this before calling the store.
type SearchSort struct {
	Field string
	Order SortOrder
}

// DateBound is an inclusive [From, To] range over a timestamp column. Either end
// is optional; a nil pointer means that bound is unconstrained.
type DateBound struct {
	From *time.Time
	To   *time.Time
}

// SearchPage holds resolved, clamped pagination for a search query. Offset and
// Limit are already defaulted and bounded by the handler before the store runs.
type SearchPage struct {
	Offset int
	Limit  int
}

// ApplicationSearchParams is the resolved input to Applications.Search. All
// filtering is optional; the zero value (empty query, no facets, no dates)
// selects everything. Follows the OffsetPage/OffsetResult convention: the store
// receives already-defaulted values and never applies request-layer policy.
type ApplicationSearchParams struct {
	Dates    map[string]DateBound
	Sort     SearchSort
	Query    string
	Statuses []string
	Page     SearchPage
}

// ApplicationSearchResult carries a page of applications alongside the metadata
// the search response needs. Mirrors OffsetResult with search extras: TotalItems
// for page math and Facets for the distinct values present in the filtered set.
type ApplicationSearchResult struct {
	Items      []*Application
	Statuses   []string
	TotalItems int64
}
