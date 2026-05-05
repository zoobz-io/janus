package models

// DefaultPageSize is the default number of results per page.
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size.
const MaxPageSize = 100

// OffsetPage holds offset-based pagination parameters for store queries.
type OffsetPage struct {
	Offset int
	Limit  int
}

// PageSize returns the effective page size, clamped to bounds.
func (p OffsetPage) PageSize() int {
	if p.Limit <= 0 {
		return DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		return MaxPageSize
	}
	return p.Limit
}

// OffsetResult holds offset-based pagination metadata alongside results.
type OffsetResult[T any] struct {
	Items  []*T
	Total  int64
	Offset int
}
