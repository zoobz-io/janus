package wire

import (
	"strings"
	"time"

	"github.com/zoobz-io/check"
)

// Shared search primitives live in this package alongside the request types that
// embed them: sentinel only records type relationships within the same package,
// so a cross-surface home would silently drop these component schemas from the
// generated OpenAPI spec. Every entity's search request reuses these types.

// DateRange is an inclusive [From, To] filter over a timestamp field. Both ends
// are optional; omit one to leave that bound open.
type DateRange struct {
	From *time.Time `json:"from,omitempty" description:"Inclusive lower bound (RFC 3339)" example:"2026-01-01T00:00:00Z"`
	To   *time.Time `json:"to,omitempty" description:"Inclusive upper bound (RFC 3339)" example:"2026-08-21T00:00:00Z"`
}

// SortSpec selects the sort field and direction for a search query.
type SortSpec struct {
	Field string `json:"field" description:"Field to sort by" example:"created_at"`
	Order string `json:"order" description:"Sort direction: asc or desc" example:"desc"`
}

// PageRequest is the requested pagination window. Both fields are optional
// pointers so an omitted value is distinguishable from an explicit one and can
// fall back to the contract defaults (page 1, size 25).
type PageRequest struct {
	Number *int `json:"number,omitempty" description:"1-based page number (default 1)" example:"1"`
	Size   *int `json:"size,omitempty" description:"Page size, 1-100 (default 25)" example:"25"`
}

// PageResponse is the pagination metadata returned with every search response.
type PageResponse struct {
	Number     int   `json:"number" description:"Current page number" example:"1"`
	Size       int   `json:"size" description:"Page size" example:"25"`
	TotalItems int64 `json:"total_items" description:"Total items across all pages" example:"57"`
	TotalPages int   `json:"total_pages" description:"Total number of pages" example:"3"`
}

// Application search field configuration. These allowlists are the per-entity
// search contract: which facet/date/sort fields are accepted, and the closed
// set of status values. Requests referencing anything outside them are rejected.
var (
	applicationFacetFields  = []string{"status"}
	applicationStatusValues = []string{"active", "inactive"}
	applicationDateFields   = []string{"created_at", "updated_at"}
	applicationSortFields   = []string{"created_at", "updated_at"}
	searchSortOrders        = []string{"asc", "desc"}
)

// Scope search field configuration. Unlike applications' status facet, the
// application facet's values are dynamic labels — they cannot be allowlisted, so
// only the facet field name is validated here; values are resolved (and silently
// dropped if unknown) in the transformer.
var (
	scopeFacetFields = []string{"application"}
	scopeDateFields  = []string{"created_at", "updated_at"}
	scopeSortFields  = []string{"created_at", "updated_at"}
)

// SearchScopesRequest is the request body for POST /scopes/search. Every key is
// optional: an empty body returns page 1, size 25, sorted updated_at desc,
// unfiltered, across all applications.
type SearchScopesRequest struct {
	Facets map[string][]string  `json:"facets,omitempty" description:"Field-scoped filters: OR within a field, AND across fields. application values are application names (labels)."`
	Dates  map[string]DateRange `json:"dates,omitempty" description:"Inclusive date-range filters keyed by field"`
	Sort   *SortSpec            `json:"sort,omitempty" description:"Sort specification (default updated_at desc)"`
	Page   *PageRequest         `json:"page,omitempty" description:"Pagination window (default page 1, size 25)"`
	Query  string               `json:"query,omitempty" description:"Case-insensitive infix match over name and description" example:"read"`
}

// Validate enforces the malformed-request rules: unknown facet/date/sort fields,
// invalid sort order, size out of [1,100], page number below 1. The application
// facet's values are NOT validated (dynamic labels) — an unknown application name
// simply matches nothing, it is not a 400.
func (r SearchScopesRequest) Validate() error {
	validations := []*check.Validation{
		check.OnlyKeys(r.Facets, scopeFacetFields, "facets"),
		check.OnlyKeys(r.Dates, scopeDateFields, "dates"),
	}
	if r.Sort != nil {
		validations = append(validations,
			check.OneOf(r.Sort.Field, scopeSortFields, "sort.field"),
			check.OneOf(strings.ToLower(r.Sort.Order), searchSortOrders, "sort.order"),
		)
	}
	if r.Page != nil {
		if r.Page.Number != nil {
			validations = append(validations, check.Min(*r.Page.Number, 1, "page.number"))
		}
		if r.Page.Size != nil {
			validations = append(validations, check.Between(*r.Page.Size, 1, 100, "page.size"))
		}
	}
	return check.All(validations...).Err()
}

// ScopeSearchResponse is the response body for POST /scopes/search.
type ScopeSearchResponse struct {
	Facets map[string][]string `json:"facets" description:"Distinct facet values present in the filtered set (application shown by name)"`
	Scopes []ScopeResponse     `json:"scopes" description:"Matching scopes for this page"`
	Page   PageResponse        `json:"page" description:"Pagination metadata"`
}

// Clone returns a deep copy of the response.
func (r ScopeSearchResponse) Clone() ScopeSearchResponse {
	c := r
	if r.Scopes != nil {
		c.Scopes = make([]ScopeResponse, len(r.Scopes))
		copy(c.Scopes, r.Scopes)
	}
	if r.Facets != nil {
		c.Facets = make(map[string][]string, len(r.Facets))
		for k, v := range r.Facets {
			vv := make([]string, len(v))
			copy(vv, v)
			c.Facets[k] = vv
		}
	}
	return c
}

// SearchApplicationsRequest is the request body for POST /applications/search.
// Every key is optional: an empty body returns page 1, size 25, sorted
// updated_at desc, unfiltered.
type SearchApplicationsRequest struct {
	Facets map[string][]string  `json:"facets,omitempty" description:"Field-scoped filters: OR within a field, AND across fields"`
	Dates  map[string]DateRange `json:"dates,omitempty" description:"Inclusive date-range filters keyed by field"`
	Sort   *SortSpec            `json:"sort,omitempty" description:"Sort specification (default updated_at desc)"`
	Page   *PageRequest         `json:"page,omitempty" description:"Pagination window (default page 1, size 25)"`
	Query  string               `json:"query,omitempty" description:"Case-insensitive infix match over name and slug" example:"acme"`
}

// Validate enforces the search contract's malformed-request rules: unknown
// facet/date/sort fields, unknown status values, invalid sort order, size out of
// [1,100], and page number below 1. Absent values are legal — they take
// defaults applied downstream — so only present fields are checked.
func (r SearchApplicationsRequest) Validate() error {
	validations := []*check.Validation{
		check.OnlyKeys(r.Facets, applicationFacetFields, "facets"),
		check.OnlyKeys(r.Dates, applicationDateFields, "dates"),
	}
	for _, status := range r.Facets["status"] {
		validations = append(validations, check.OneOf(status, applicationStatusValues, "facets.status"))
	}
	if r.Sort != nil {
		validations = append(validations,
			check.OneOf(r.Sort.Field, applicationSortFields, "sort.field"),
			check.OneOf(strings.ToLower(r.Sort.Order), searchSortOrders, "sort.order"),
		)
	}
	if r.Page != nil {
		if r.Page.Number != nil {
			validations = append(validations, check.Min(*r.Page.Number, 1, "page.number"))
		}
		if r.Page.Size != nil {
			validations = append(validations, check.Between(*r.Page.Size, 1, 100, "page.size"))
		}
	}
	return check.All(validations...).Err()
}

// ApplicationSearchResponse is the response body for POST /applications/search.
type ApplicationSearchResponse struct {
	Facets       map[string][]string   `json:"facets" description:"Distinct facet values present in the filtered set"`
	Applications []ApplicationResponse `json:"applications" description:"Matching applications for this page"`
	Page         PageResponse          `json:"page" description:"Pagination metadata"`
}

// Clone returns a deep copy of the response.
func (r ApplicationSearchResponse) Clone() ApplicationSearchResponse {
	c := r
	if r.Applications != nil {
		c.Applications = make([]ApplicationResponse, len(r.Applications))
		copy(c.Applications, r.Applications)
	}
	if r.Facets != nil {
		c.Facets = make(map[string][]string, len(r.Facets))
		for k, v := range r.Facets {
			vv := make([]string, len(v))
			copy(vv, v)
			c.Facets[k] = vv
		}
	}
	return c
}
