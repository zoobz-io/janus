package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/events"
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
	events.ApplicationCreated.Emit(r, events.ApplicationEvent{ApplicationID: app.ID, Name: app.Name})
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

	params, number, size := transformers.ResolveApplicationSearch(r.Body)
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

var updateApplication = rocco.PATCH[wire.UpdateApplicationRequest, wire.ApplicationResponse]("/applications/{app_id}", func(r *rocco.Request[wire.UpdateApplicationRequest]) (wire.ApplicationResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	app, err := apps.Update(r, pathID(r.Params, "app_id"), r.Body.Name, r.Body.Status)
	if err != nil {
		return wire.ApplicationResponse{}, ErrApplicationNotFound
	}
	events.ApplicationUpdated.Emit(r, events.ApplicationEvent{ApplicationID: app.ID, Name: app.Name})
	return transformers.ApplicationToResponse(app), nil
}).
	WithSummary("Update an application").
	WithDescription("Updates an application's name and status. Set status to 'inactive' to soft-disable it — there is no hard delete.").
	WithTags("Applications").
	WithPathParams("app_id").
	WithAuthentication().
	WithErrors(ErrApplicationNotFound, rocco.ErrValidationFailed)
