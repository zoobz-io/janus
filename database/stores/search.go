package stores

import (
	"context"
	"sort"
	"time"

	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

// runSearch executes the standard three-query search every entity's Search
// shares: a filtered page (ordered by the requested sort with an id tiebreak,
// then windowed), the total count, and the distinct facet. base returns a fresh
// filtered query (WHERE already applied via the searchFilterable helper) — it is
// called once for the page and once for the facet; count is the same WHERE on an
// aggregate. facetOf reads the facet column off a row.
func runSearch[M any](
	ctx context.Context,
	params map[string]any,
	base func() *soy.Query[M],
	count *soy.Aggregate[M],
	srt SearchSort,
	offset, limit int,
	facetField string,
	facetOf func(*M) string,
) (items []*M, total int64, facetValues []string, err error) {
	items, err = base().
		OrderBy(srt.Field, srt.Order).
		OrderBy("id", SortAsc).
		Limit(limit).
		Offset(offset).
		Exec(ctx, params)
	if err != nil {
		return nil, 0, nil, err
	}
	t, err := count.Exec(ctx, params)
	if err != nil {
		return nil, 0, nil, err
	}
	rows, err := base().Fields(facetField).Distinct().Exec(ctx, params)
	if err != nil {
		return nil, 0, nil, err
	}
	facetValues = make([]string, 0, len(rows))
	for _, row := range rows {
		facetValues = append(facetValues, facetOf(row))
	}
	sort.Strings(facetValues)
	return items, int64(t), facetValues, nil
}

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

// ScopeSearchParams is the resolved input to Scopes.Search. All filtering is
// optional; the zero value selects everything.
type ScopeSearchParams struct {
	// Dates maps a timestamp column ("created_at"/"updated_at") to its inclusive
	// bounds. Absent keys are unconstrained.
	Dates map[string]DateBound
	// Sort is the resolved ordering (defaulted to updated_at DESC by the handler).
	Sort SearchSort
	// Query is the case-insensitive infix text match applied across name/description.
	Query string
	// ApplicationIDs is the application facet, already resolved from labels to ids.
	// nil means no application filter; a non-nil slice (INCLUDING an empty one)
	// means filter to exactly these ids — so a facet whose names all failed to
	// resolve matches nothing, rather than silently disabling the filter.
	ApplicationIDs []string
	// Page is the resolved, clamped pagination window.
	Page SearchPage
}

// ScopeSearchResult carries a page of scopes plus TotalItems for page math and
// ApplicationIDs — the distinct application ids present in the filtered set,
// which the transformer resolves to names for the facet.
type ScopeSearchResult struct {
	Items          []*models.Scope
	ApplicationIDs []string
	TotalItems     int64
}
