package stores

import (
	"context"
	"sort"
	"time"

	"github.com/lib/pq"
	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

// Shared WHERE-clause primitives for the search template. Each is generic over
// the builder type via searchFilterable, so it composes onto both the page query
// and the count aggregate. Entities' apply<Entity>Search functions assemble these.

// applyTextSearch ORs an escaped infix ILIKE of query across the given fields.
// An empty query adds nothing.
func applyTextSearch[B searchFilterable[B]](b B, query string, params map[string]any, fields ...string) B {
	if query == "" {
		return b
	}
	params["query"] = "%" + escapeLike(query) + "%"
	conds := make([]soy.Condition, len(fields))
	for i, f := range fields {
		conds[i] = soy.C(f, "ILIKE", "query")
	}
	return b.WhereOr(conds...)
}

// applyStatusFacet filters status to the OR-set via IN (rendered = ANY). An
// empty set adds nothing.
func applyStatusFacet[B searchFilterable[B]](b B, statuses []string, params map[string]any) B {
	if len(statuses) == 0 {
		return b
	}
	params["statuses"] = pq.Array(statuses)
	return b.Where("status", "IN", "statuses")
}

// applyDateBounds adds inclusive >=/<= filters for each named timestamp column
// that has a bound in dates; each end is optional.
func applyDateBounds[B searchFilterable[B]](b B, dates map[string]DateBound, fields []string, params map[string]any) B {
	for _, field := range fields {
		bound, ok := dates[field]
		if !ok {
			continue
		}
		if bound.From != nil {
			param := field + "_from"
			params[param] = *bound.From
			b = b.Where(field, ">=", param)
		}
		if bound.To != nil {
			param := field + "_to"
			params[param] = *bound.To
			b = b.Where(field, "<=", param)
		}
	}
	return b
}

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

// UserSearchParams is the resolved input to Users.Search. All filtering is
// optional; the zero value selects everything.
type UserSearchParams struct {
	// Dates maps a timestamp column ("created_at"/"updated_at") to its inclusive
	// bounds. Absent keys are unconstrained.
	Dates map[string]DateBound
	// Sort is the resolved ordering (defaulted to updated_at DESC by the handler).
	Sort SearchSort
	// Query is the case-insensitive infix text match over email/display_name.
	Query string
	// Statuses is the OR-set for the status facet. Empty means no status filter.
	Statuses []string
	// Page is the resolved, clamped pagination window.
	Page SearchPage
}

// UserSearchResult carries a page of users plus TotalItems for page math and
// Statuses — the distinct facet values present in the filtered set.
type UserSearchResult struct {
	Items      []*models.User
	Statuses   []string
	TotalItems int64
}

// TenantSearchParams is the resolved input to Tenants.Search. All filtering is
// optional; the zero value selects everything.
type TenantSearchParams struct {
	// Dates maps a timestamp column ("created_at"/"updated_at") to its inclusive
	// bounds. Absent keys are unconstrained.
	Dates map[string]DateBound
	// Sort is the resolved ordering (defaulted to updated_at DESC by the handler).
	Sort SearchSort
	// Query is the case-insensitive infix text match over name/slug.
	Query string
	// Statuses is the OR-set for the status facet. Empty means no status filter.
	Statuses []string
	// Page is the resolved, clamped pagination window.
	Page SearchPage
}

// TenantSearchResult carries a page of tenants plus TotalItems for page math and
// Statuses — the distinct facet values present in the filtered set.
type TenantSearchResult struct {
	Items      []*models.Tenant
	Statuses   []string
	TotalItems int64
}
