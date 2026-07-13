package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listScopes = rocco.GET[rocco.NoBody, wire.ScopeListResponse]("/applications/{app_id}/scopes", func(r *rocco.Request[rocco.NoBody]) (wire.ScopeListResponse, error) {
	scopes := sum.MustUse[contracts.Scopes](r)
	list, err := scopes.ListByApplication(r, pathID(r.Params, "app_id"))
	if err != nil {
		return wire.ScopeListResponse{}, err
	}
	return wire.ScopeListResponse{Scopes: transformers.ScopesToResponse(list)}, nil
}).
	WithSummary("List application scopes").
	WithTags("Scopes").
	WithPathParams("app_id").
	WithAuthentication()

var createScope = rocco.POST[wire.CreateScopeRequest, wire.ScopeResponse]("/applications/{app_id}/scopes", func(r *rocco.Request[wire.CreateScopeRequest]) (wire.ScopeResponse, error) {
	scopes := sum.MustUse[contracts.Scopes](r)
	sc, err := scopes.Define(r, pathID(r.Params, "app_id"), r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.ScopeResponse{}, err
	}
	return transformers.ScopeToResponse(sc), nil
}).
	WithSummary("Define an application scope").
	WithTags("Scopes").
	WithPathParams("app_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var deleteScope = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/applications/{app_id}/scopes/{scope_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	scopes := sum.MustUse[contracts.Scopes](r)
	if err := scopes.Delete(r, pathID(r.Params, "scope_id")); err != nil {
		return rocco.NoBody{}, ErrScopeNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete an application scope").
	WithTags("Scopes").
	WithPathParams("app_id", "scope_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrScopeNotFound)
