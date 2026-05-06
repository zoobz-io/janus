package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// ---------------------------------------------------------------------------
// Application catalog (any authenticated user)
// ---------------------------------------------------------------------------

var listApplications = rocco.GET[rocco.NoBody, wire.ApplicationListResponse]("/applications", func(r *rocco.Request[rocco.NoBody]) (wire.ApplicationListResponse, error) {
	store := sum.MustUse[contracts.Applications](r)
	apps, err := store.ListApplications(r)
	if err != nil {
		return wire.ApplicationListResponse{}, err
	}
	return wire.ApplicationListResponse{
		Applications: transformers.ApplicationsToResponse(apps),
	}, nil
}).
	WithSummary("List available applications").
	WithTags("Applications").
	WithAuthentication()

// ---------------------------------------------------------------------------
// Tenant application authorization (admin/owner)
// ---------------------------------------------------------------------------

var listTenantApplications = rocco.GET[rocco.NoBody, wire.TenantApplicationListResponse]("/tenants/{tenant_id}/applications", func(r *rocco.Request[rocco.NoBody]) (wire.TenantApplicationListResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleViewer, models.UserRoleEditor, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.TenantApplicationListResponse{}, err
	}

	store := sum.MustUse[contracts.TenantApplications](r)
	tas, err := store.ListByTenant(r, tenantID)
	if err != nil {
		return wire.TenantApplicationListResponse{}, err
	}
	return wire.TenantApplicationListResponse{
		Authorizations: transformers.TenantApplicationsToResponse(tas),
	}, nil
}).
	WithSummary("List applications authorized for a tenant").
	WithTags("Applications").
	WithPathParams("tenant_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound)

var authorizeTenantApplication = rocco.POST[wire.AuthorizeApplicationRequest, wire.TenantApplicationResponse]("/tenants/{tenant_id}/applications", func(r *rocco.Request[wire.AuthorizeApplicationRequest]) (wire.TenantApplicationResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.TenantApplicationResponse{}, err
	}

	taStore := sum.MustUse[contracts.TenantApplications](r)
	ta, err := taStore.Authorize(r, tenantID, r.Body.ApplicationID)
	if err != nil {
		return wire.TenantApplicationResponse{}, err
	}
	return transformers.TenantApplicationToResponse(ta), nil
}).
	WithSummary("Authorize a tenant to use an application").
	WithTags("Applications").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, rocco.ErrValidationFailed)

var revokeTenantApplication = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tenants/{tenant_id}/applications/{application_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tenantID := pathID(r.Params, "tenant_id")
	applicationID := pathID(r.Params, "application_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return rocco.NoBody{}, err
	}

	taStore := sum.MustUse[contracts.TenantApplications](r)
	if err := taStore.Revoke(r, tenantID, applicationID); err != nil {
		return rocco.NoBody{}, ErrApplicationNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Revoke a tenant's access to an application").
	WithTags("Applications").
	WithPathParams("tenant_id", "application_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, ErrApplicationNotFound)

// ---------------------------------------------------------------------------
// User application grants (admin/owner)
// ---------------------------------------------------------------------------

var listUserApplications = rocco.GET[rocco.NoBody, wire.UserApplicationListResponse]("/tenants/{tenant_id}/users/{user_id}/applications", func(r *rocco.Request[rocco.NoBody]) (wire.UserApplicationListResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleViewer, models.UserRoleEditor, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.UserApplicationListResponse{}, err
	}

	store := sum.MustUse[contracts.UserApplications](r)
	uas, err := store.ListByUser(r, userID, tenantID)
	if err != nil {
		return wire.UserApplicationListResponse{}, err
	}
	return wire.UserApplicationListResponse{
		Grants: transformers.UserApplicationsToResponse(uas),
	}, nil
}).
	WithSummary("List applications granted to a user within a tenant").
	WithTags("Applications").
	WithPathParams("tenant_id", "user_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound)

var grantUserApplication = rocco.POST[wire.GrantApplicationRequest, wire.UserApplicationResponse]("/tenants/{tenant_id}/applications/grants", func(r *rocco.Request[wire.GrantApplicationRequest]) (wire.UserApplicationResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.UserApplicationResponse{}, err
	}

	uaStore := sum.MustUse[contracts.UserApplications](r)
	ua, err := uaStore.Grant(r, r.Body.UserID, tenantID, r.Body.ApplicationID)
	if err != nil {
		return wire.UserApplicationResponse{}, err
	}
	return transformers.UserApplicationToResponse(ua), nil
}).
	WithSummary("Grant a user access to an application within a tenant").
	WithTags("Applications").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, rocco.ErrValidationFailed)

var revokeUserApplication = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tenants/{tenant_id}/users/{user_id}/applications/{application_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")
	applicationID := pathID(r.Params, "application_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return rocco.NoBody{}, err
	}

	uaStore := sum.MustUse[contracts.UserApplications](r)
	if err := uaStore.Revoke(r, userID, tenantID, applicationID); err != nil {
		return rocco.NoBody{}, ErrApplicationNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Revoke a user's access to an application within a tenant").
	WithTags("Applications").
	WithPathParams("tenant_id", "user_id", "application_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, ErrApplicationNotFound)

// ---------------------------------------------------------------------------
// My application grants (self-service)
// ---------------------------------------------------------------------------

var listMyApplications = rocco.GET[rocco.NoBody, wire.UserApplicationListResponse]("/me/tenants/{tenant_id}/applications", func(r *rocco.Request[rocco.NoBody]) (wire.UserApplicationListResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	store := sum.MustUse[contracts.UserApplications](r)
	uas, err := store.ListByUser(r, r.Identity.ID(), tenantID)
	if err != nil {
		return wire.UserApplicationListResponse{}, err
	}
	return wire.UserApplicationListResponse{
		Grants: transformers.UserApplicationsToResponse(uas),
	}, nil
}).
	WithSummary("List my application grants within a tenant").
	WithTags("Applications").
	WithPathParams("tenant_id").
	WithAuthentication()
