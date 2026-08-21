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

// validateSearchRequest runs the search contract's shared malformed-request
// checks for an entity whose status facet is a closed enum: facet/date field
// allowlists, status-value enum, sort field/order, and page bounds. Absent
// values are legal (they take downstream defaults), so only present fields are
// checked.
func validateSearchRequest(
	facets map[string][]string,
	dates map[string]DateRange,
	sort *SortSpec,
	page *PageRequest,
	facetFields, statusValues, dateFields, sortFields []string,
) error {
	validations := []*check.Validation{
		check.OnlyKeys(facets, facetFields, "facets"),
		check.OnlyKeys(dates, dateFields, "dates"),
	}
	for _, status := range facets["status"] {
		validations = append(validations, check.OneOf(status, statusValues, "facets.status"))
	}
	if sort != nil {
		validations = append(validations,
			check.OneOf(sort.Field, sortFields, "sort.field"),
			check.OneOf(strings.ToLower(sort.Order), searchSortOrders, "sort.order"),
		)
	}
	if page != nil {
		if page.Number != nil {
			validations = append(validations, check.Min(*page.Number, 1, "page.number"))
		}
		if page.Size != nil {
			validations = append(validations, check.Between(*page.Size, 1, 100, "page.size"))
		}
	}
	return check.All(validations...).Err()
}

// User search field configuration. status is a closed enum (like applications),
// so its facet values are wire-validated.
var (
	userFacetFields  = []string{"status"}
	userStatusValues = []string{"active", "inactive"}
	userDateFields   = []string{"created_at", "updated_at"}
	userSortFields   = []string{"created_at", "updated_at"}
)

// SearchUsersRequest is the request body for POST /users/search. Every key is
// optional: an empty body returns page 1, size 25, sorted updated_at desc,
// unfiltered.
type SearchUsersRequest struct {
	Facets map[string][]string  `json:"facets,omitempty" description:"Field-scoped filters: OR within a field, AND across fields"`
	Dates  map[string]DateRange `json:"dates,omitempty" description:"Inclusive date-range filters keyed by field"`
	Sort   *SortSpec            `json:"sort,omitempty" description:"Sort specification (default updated_at desc)"`
	Page   *PageRequest         `json:"page,omitempty" description:"Pagination window (default page 1, size 25)"`
	Query  string               `json:"query,omitempty" description:"Case-insensitive infix match over email and display name" example:"jane"`
}

// Validate enforces the malformed-request rules: unknown facet/date/sort fields,
// unknown status values, invalid sort order, size out of [1,100], page number
// below 1.
func (r SearchUsersRequest) Validate() error {
	return validateSearchRequest(r.Facets, r.Dates, r.Sort, r.Page,
		userFacetFields, userStatusValues, userDateFields, userSortFields)
}

// UserSearchResponse is the response body for POST /users/search.
type UserSearchResponse struct {
	Facets map[string][]string `json:"facets" description:"Distinct facet values present in the filtered set"`
	Users  []UserResponse      `json:"users" description:"Matching users for this page"`
	Page   PageResponse        `json:"page" description:"Pagination metadata"`
}

// Clone returns a deep copy of the response.
func (r UserSearchResponse) Clone() UserSearchResponse {
	c := r
	if r.Users != nil {
		c.Users = make([]UserResponse, len(r.Users))
		copy(c.Users, r.Users)
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

// Tenant search field configuration. status is a closed enum (active/suspended),
// so its facet values are wire-validated.
var (
	tenantFacetFields  = []string{"status"}
	tenantStatusValues = []string{"active", "suspended"}
	tenantDateFields   = []string{"created_at", "updated_at"}
	tenantSortFields   = []string{"created_at", "updated_at"}
)

// SearchTenantsRequest is the request body for POST /tenants/search. Every key is
// optional: an empty body returns page 1, size 25, sorted updated_at desc,
// unfiltered.
type SearchTenantsRequest struct {
	Facets map[string][]string  `json:"facets,omitempty" description:"Field-scoped filters: OR within a field, AND across fields"`
	Dates  map[string]DateRange `json:"dates,omitempty" description:"Inclusive date-range filters keyed by field"`
	Sort   *SortSpec            `json:"sort,omitempty" description:"Sort specification (default updated_at desc)"`
	Page   *PageRequest         `json:"page,omitempty" description:"Pagination window (default page 1, size 25)"`
	Query  string               `json:"query,omitempty" description:"Case-insensitive infix match over name and slug" example:"acme"`
}

// Validate enforces the malformed-request rules: unknown facet/date/sort fields,
// unknown status values, invalid sort order, size out of [1,100], page number
// below 1.
func (r SearchTenantsRequest) Validate() error {
	return validateSearchRequest(r.Facets, r.Dates, r.Sort, r.Page,
		tenantFacetFields, tenantStatusValues, tenantDateFields, tenantSortFields)
}

// TenantSearchResponse is the response body for POST /tenants/search.
type TenantSearchResponse struct {
	Facets  map[string][]string `json:"facets" description:"Distinct facet values present in the filtered set"`
	Tenants []TenantResponse    `json:"tenants" description:"Matching tenants for this page"`
	Page    PageResponse        `json:"page" description:"Pagination metadata"`
}

// Clone returns a deep copy of the response.
func (r TenantSearchResponse) Clone() TenantSearchResponse {
	c := r
	if r.Tenants != nil {
		c.Tenants = make([]TenantResponse, len(r.Tenants))
		copy(c.Tenants, r.Tenants)
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
	return validateSearchRequest(r.Facets, r.Dates, r.Sort, r.Page,
		applicationFacetFields, applicationStatusValues, applicationDateFields, applicationSortFields)
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
