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
	WithTags("Tenants").
	WithQueryParams("limit", "offset").
	WithAuthentication()

var getTenant = rocco.GET[rocco.NoBody, wire.TenantResponse]("/tenants/{tenant_id}", func(r *rocco.Request[rocco.NoBody]) (wire.TenantResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	tenant, err := tenants.GetTenant(r, pathID(r.Params, "tenant_id"))
	if err != nil {
		return wire.TenantResponse{}, ErrTenantNotFound
	}
	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Get a tenant").
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
	WithTags("Tenants").
	WithPathParams("tenant_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound, rocco.ErrValidationFailed)
