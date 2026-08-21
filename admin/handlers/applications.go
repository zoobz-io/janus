package handlers

import (
	"strings"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// Search contract defaults, applied when the request omits the field.
const (
	defaultSearchPageNumber = 1
	defaultSearchPageSize   = 25
	defaultSearchSortField  = "updated_at"
	defaultSearchSortOrder  = models.SortDesc
)

var listApplications = rocco.GET[rocco.NoBody, wire.ApplicationListResponse]("/applications", func(r *rocco.Request[rocco.NoBody]) (wire.ApplicationListResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	list, err := apps.ListAll(r)
	if err != nil {
		return wire.ApplicationListResponse{}, err
	}
	return wire.ApplicationListResponse{Applications: transformers.ApplicationsToResponse(list)}, nil
}).
	WithSummary("List all applications").
	WithDescription("Returns every application on the mesh, including inactive ones. The public catalog only exposes active applications.").
	WithTags("Applications").
	WithAuthentication()

var getApplication = rocco.GET[rocco.NoBody, wire.ApplicationResponse]("/applications/{app_id}", func(r *rocco.Request[rocco.NoBody]) (wire.ApplicationResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	app, err := apps.GetApplication(r, pathID(r.Params, "app_id"))
	if err != nil {
		return wire.ApplicationResponse{}, ErrApplicationNotFound
	}
	return transformers.ApplicationToResponse(app), nil
}).
	WithSummary("Get an application").
	WithDescription("Fetches a single application by ID, including its current status.").
	WithTags("Applications").
	WithPathParams("app_id").
	WithAuthentication().
	WithErrors(ErrApplicationNotFound)

var createApplication = rocco.POST[wire.CreateApplicationRequest, wire.ApplicationResponse]("/applications", func(r *rocco.Request[wire.CreateApplicationRequest]) (wire.ApplicationResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	app, err := apps.CreateApplication(r, r.Body.Name, r.Body.Slug)
	if err != nil {
		return wire.ApplicationResponse{}, err
	}
	return transformers.ApplicationToResponse(app), nil
}).
	WithSummary("Register an application").
	WithDescription("Registers a new application. The slug must match the certificate CN of the service's mesh nodes.").
	WithTags("Applications").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var searchApplications = rocco.POST[wire.SearchApplicationsRequest, wire.ApplicationSearchResponse]("/applications/search", func(r *rocco.Request[wire.SearchApplicationsRequest]) (wire.ApplicationSearchResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)

	params, number, size := resolveApplicationSearch(r.Body)
	result, err := apps.Search(r, params)
	if err != nil {
		return wire.ApplicationSearchResponse{}, err
	}
	return transformers.ApplicationSearchToResponse(result, number, size), nil
}).
	WithSummary("Search applications").
	WithDescription("Query applications with text search over name and slug, status faceting, date-range filters, sorting, and pagination. An empty body returns page 1, size 25, sorted updated_at desc, unfiltered.").
	WithTags("Applications").
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

// resolveApplicationSearch turns a validated search request into store params,
// applying every contract default in one place: page 1, size 25, sort
// updated_at desc. It returns the params plus the resolved page number and size
// so the transformer can build the page metadata. The request is assumed valid
// (rocco ran Validate first), so number >= 1 and size in [1,100].
func resolveApplicationSearch(body wire.SearchApplicationsRequest) (models.ApplicationSearchParams, int, int) {
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

	params := models.ApplicationSearchParams{
		Query:    body.Query,
		Statuses: body.Facets["status"],
		Dates:    searchDateBounds(body.Dates),
		Sort:     models.SearchSort{Field: sortField, Order: sortOrder},
		Page:     models.SearchPage{Offset: (number - 1) * size, Limit: size},
	}
	return params, number, size
}

// sortOrderToSQL maps a request sort order ("asc"/"desc", case-insensitive) to
// the SQL direction. Unrecognized values never reach here — Validate rejects
// them — so anything non-asc defaults to descending.
func sortOrderToSQL(order string) models.SortOrder {
	if strings.EqualFold(order, "asc") {
		return models.SortAsc
	}
	return models.SortDesc
}

// searchDateBounds converts wire date ranges into store date bounds.
func searchDateBounds(in map[string]wire.DateRange) map[string]models.DateBound {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]models.DateBound, len(in))
	for field, dr := range in {
		out[field] = models.DateBound{From: dr.From, To: dr.To}
	}
	return out
}

var updateApplication = rocco.PATCH[wire.UpdateApplicationRequest, wire.ApplicationResponse]("/applications/{app_id}", func(r *rocco.Request[wire.UpdateApplicationRequest]) (wire.ApplicationResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	app, err := apps.Update(r, pathID(r.Params, "app_id"), r.Body.Name, r.Body.Status)
	if err != nil {
		return wire.ApplicationResponse{}, ErrApplicationNotFound
	}
	return transformers.ApplicationToResponse(app), nil
}).
	WithSummary("Update an application").
	WithDescription("Updates an application's name and status. Set status to 'inactive' to soft-disable it — there is no hard delete.").
	WithTags("Applications").
	WithPathParams("app_id").
	WithAuthentication().
	WithErrors(ErrApplicationNotFound, rocco.ErrValidationFailed)
