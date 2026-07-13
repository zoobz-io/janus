package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
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
	WithTags("Applications").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateApplication = rocco.PATCH[wire.UpdateApplicationRequest, wire.ApplicationResponse]("/applications/{app_id}", func(r *rocco.Request[wire.UpdateApplicationRequest]) (wire.ApplicationResponse, error) {
	apps := sum.MustUse[contracts.Applications](r)
	app, err := apps.Update(r, pathID(r.Params, "app_id"), r.Body.Name, r.Body.Status)
	if err != nil {
		return wire.ApplicationResponse{}, ErrApplicationNotFound
	}
	return transformers.ApplicationToResponse(app), nil
}).
	WithSummary("Update an application").
	WithTags("Applications").
	WithPathParams("app_id").
	WithAuthentication().
	WithErrors(ErrApplicationNotFound, rocco.ErrValidationFailed)
