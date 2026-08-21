package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listTenants = rocco.GET[rocco.NoBody, wire.TenantListResponse]("/tenants", func(r *rocco.Request[rocco.NoBody]) (wire.TenantListResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	result, err := tenants.ListTenants(r, offsetPage(r.Params))
	if err != nil {
		return wire.TenantListResponse{}, err
	}
	return wire.TenantListResponse{Tenants: transformers.TenantsToResponse(result.Items), Total: result.Total}, nil
}).
	WithSummary("List tenants").
	WithDescription("Returns a paginated list of all tenants. Use limit and offset to page.").
	WithTags("Tenants").
	WithQueryParams("limit", "offset").
	WithAuthentication()

var searchTenants = rocco.POST[wire.SearchTenantsRequest, wire.TenantSearchResponse]("/tenants/search", func(r *rocco.Request[wire.SearchTenantsRequest]) (wire.TenantSearchResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)

	params, number, size := transformers.ResolveTenantSearch(r.Body)
	result, err := tenants.Search(r, params)
	if err != nil {
		return wire.TenantSearchResponse{}, err
	}
	return transformers.TenantSearchToResponse(result, number, size), nil
}).
	WithSummary("Search tenants").
	WithDescription("Query tenants with text search over name and slug, status faceting, date-range filters, sorting, and pagination. An empty body returns page 1, size 25, sorted updated_at desc, unfiltered.").
	WithTags("Tenants").
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var getTenant = rocco.GET[rocco.NoBody, wire.TenantResponse]("/tenants/{tenant_id}", func(r *rocco.Request[rocco.NoBody]) (wire.TenantResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	tenant, err := tenants.GetTenant(r, pathID(r.Params, "tenant_id"))
	if err != nil {
		return wire.TenantResponse{}, ErrTenantNotFound
	}
	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Get a tenant").
	WithDescription("Fetches a single tenant by ID.").
	WithTags("Tenants").
	WithPathParams("tenant_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound)

var createTenant = rocco.POST[wire.CreateTenantRequest, wire.TenantResponse]("/tenants", func(r *rocco.Request[wire.CreateTenantRequest]) (wire.TenantResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	tenant, err := tenants.CreateTenant(r, r.Body.Name, r.Body.Slug)
	if err != nil {
		return wire.TenantResponse{}, err
	}
	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Create a tenant").
	WithDescription("Creates a new customer organization.").
	WithTags("Tenants").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateTenant = rocco.PATCH[wire.UpdateTenantRequest, wire.TenantResponse]("/tenants/{tenant_id}", func(r *rocco.Request[wire.UpdateTenantRequest]) (wire.TenantResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	tenant, err := tenants.UpdateTenant(r, pathID(r.Params, "tenant_id"), r.Body.Name, r.Body.Status)
	if err != nil {
		return wire.TenantResponse{}, ErrTenantNotFound
	}
	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Update a tenant").
	WithDescription("Updates a tenant's name and status. Set status to 'suspended' to disable it.").
	WithTags("Tenants").
	WithPathParams("tenant_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound, rocco.ErrValidationFailed)
