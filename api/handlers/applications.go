package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// Application catalog — any authenticated user can browse.

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

// Self-service — what apps do I have access to in a tenant?

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
