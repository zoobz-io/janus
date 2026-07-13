package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listGrants = rocco.GET[rocco.NoBody, wire.GrantListResponse]("/applications/{app_id}/grants", func(r *rocco.Request[rocco.NoBody]) (wire.GrantListResponse, error) {
	grants := sum.MustUse[contracts.Grants](r)
	list, err := grants.ListByTenantAndApp(r, r.Params.Query["tenant_id"], pathID(r.Params, "app_id"))
	if err != nil {
		return wire.GrantListResponse{}, err
	}
	return wire.GrantListResponse{Grants: transformers.GrantsToResponse(list)}, nil
}).
	WithSummary("List user grants for an application in a tenant").
	WithDescription("Returns the per-user grants for an application within a tenant. Pass tenant_id as a query parameter.").
	WithTags("Grants").
	WithPathParams("app_id").
	WithQueryParams("tenant_id").
	WithAuthentication()

var createGrant = rocco.POST[wire.CreateGrantRequest, wire.GrantResponse]("/applications/{app_id}/grants", func(r *rocco.Request[wire.CreateGrantRequest]) (wire.GrantResponse, error) {
	grants := sum.MustUse[contracts.Grants](r)
	appID := pathID(r.Params, "app_id")

	g, err := grants.Grant(r, r.Body.UserID, r.Body.TenantID, appID, r.Body.Roles, r.Body.Scopes)
	if err != nil {
		return wire.GrantResponse{}, err
	}
	if r.Body.TierID != "" {
		g, err = grants.SetTier(r, r.Body.UserID, r.Body.TenantID, appID, r.Body.TierID)
		if err != nil {
			return wire.GrantResponse{}, err
		}
	}
	return transformers.GrantToResponse(g), nil
}).
	WithSummary("Grant a user access to an application").
	WithDescription("Grants a user access to an application within a tenant, with explicit roles/scopes and an optional tier. Explicit scopes are admin-assigned; tier scopes are inherited.").
	WithTags("Grants").
	WithPathParams("app_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateGrant = rocco.PATCH[wire.UpdateGrantRequest, wire.GrantResponse]("/applications/{app_id}/grants/{tenant_id}/{user_id}", func(r *rocco.Request[wire.UpdateGrantRequest]) (wire.GrantResponse, error) {
	grants := sum.MustUse[contracts.Grants](r)
	appID := pathID(r.Params, "app_id")
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")

	if _, err := grants.UpdateAccess(r, userID, tenantID, appID, r.Body.Roles, r.Body.Scopes); err != nil {
		return wire.GrantResponse{}, err
	}
	g, err := grants.SetTier(r, userID, tenantID, appID, r.Body.TierID)
	if err != nil {
		return wire.GrantResponse{}, err
	}
	return transformers.GrantToResponse(g), nil
}).
	WithSummary("Update a user's grant").
	WithDescription("Replaces a grant's roles and scopes and sets (or clears) its tier.").
	WithTags("Grants").
	WithPathParams("app_id", "tenant_id", "user_id").
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var revokeGrant = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/applications/{app_id}/grants/{tenant_id}/{user_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	grants := sum.MustUse[contracts.Grants](r)
	if err := grants.Revoke(r, pathID(r.Params, "user_id"), pathID(r.Params, "tenant_id"), pathID(r.Params, "app_id")); err != nil {
		return rocco.NoBody{}, ErrUserNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Revoke a user's grant").
	WithDescription("Removes a user's access to an application within a tenant.").
	WithTags("Grants").
	WithPathParams("app_id", "tenant_id", "user_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrUserNotFound)
