package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listLicenses = rocco.GET[rocco.NoBody, wire.LicenseListResponse]("/applications/{app_id}/licenses", func(r *rocco.Request[rocco.NoBody]) (wire.LicenseListResponse, error) {
	licenses := sum.MustUse[contracts.Licenses](r)
	list, err := licenses.ListByApplication(r, pathID(r.Params, "app_id"))
	if err != nil {
		return wire.LicenseListResponse{}, err
	}
	return wire.LicenseListResponse{Licenses: transformers.LicensesToResponse(list)}, nil
}).
	WithSummary("List tenants licensed for an application").
	WithTags("Licenses").
	WithPathParams("app_id").
	WithAuthentication()

var authorizeLicense = rocco.POST[wire.AuthorizeLicenseRequest, wire.LicenseResponse]("/applications/{app_id}/licenses", func(r *rocco.Request[wire.AuthorizeLicenseRequest]) (wire.LicenseResponse, error) {
	licenses := sum.MustUse[contracts.Licenses](r)
	lic, err := licenses.Authorize(r, r.Body.TenantID, pathID(r.Params, "app_id"))
	if err != nil {
		return wire.LicenseResponse{}, err
	}
	return transformers.LicenseToResponse(lic), nil
}).
	WithSummary("Authorize a tenant for an application").
	WithTags("Licenses").
	WithPathParams("app_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var revokeLicense = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/applications/{app_id}/licenses/{tenant_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	licenses := sum.MustUse[contracts.Licenses](r)
	if err := licenses.Revoke(r, pathID(r.Params, "tenant_id"), pathID(r.Params, "app_id")); err != nil {
		return rocco.NoBody{}, ErrTenantNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Revoke a tenant's license for an application").
	WithTags("Licenses").
	WithPathParams("app_id", "tenant_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrTenantNotFound)
