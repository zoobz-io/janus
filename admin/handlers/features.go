package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listFeatures = rocco.GET[rocco.NoBody, wire.FeatureListResponse]("/applications/{app_id}/tiers/{tier_id}/features", func(r *rocco.Request[rocco.NoBody]) (wire.FeatureListResponse, error) {
	features := sum.MustUse[contracts.Features](r)
	list, err := features.ListByTier(r, pathID(r.Params, "tier_id"))
	if err != nil {
		return wire.FeatureListResponse{}, err
	}
	return wire.FeatureListResponse{Features: transformers.FeaturesToResponse(list)}, nil
}).
	WithName("list-features").
	WithSummary("List scopes bundled into a tier").
	WithScopes("directory:read").
	WithDescription("Returns the scopes bundled into a tier — the features every subscriber on this tier inherits.").
	WithTags("Tiers").
	WithPathParams("app_id", "tier_id").
	WithAuthentication()

var addFeature = rocco.POST[wire.AddFeatureRequest, wire.FeatureResponse]("/applications/{app_id}/tiers/{tier_id}/features", func(r *rocco.Request[wire.AddFeatureRequest]) (wire.FeatureResponse, error) {
	features := sum.MustUse[contracts.Features](r)
	f, err := features.Add(r, pathID(r.Params, "tier_id"), r.Body.ScopeID)
	if err != nil {
		return wire.FeatureResponse{}, err
	}
	return transformers.FeatureToResponse(f), nil
}).
	WithName("add-feature").
	WithSummary("Bundle a scope into a tier").
	WithScopes("applications:manage").
	WithDescription("Adds a scope to a tier's bundle. Users on the tier inherit it immediately — scope composition is dynamic.").
	WithTags("Tiers").
	WithPathParams("app_id", "tier_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var removeFeature = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/applications/{app_id}/tiers/{tier_id}/features/{scope_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	features := sum.MustUse[contracts.Features](r)
	if err := features.Remove(r, pathID(r.Params, "tier_id"), pathID(r.Params, "scope_id")); err != nil {
		return rocco.NoBody{}, ErrScopeNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithName("remove-feature").
	WithSummary("Unbundle a scope from a tier").
	WithScopes("applications:manage").
	WithDescription("Removes a scope from a tier's bundle.").
	WithTags("Tiers").
	WithPathParams("app_id", "tier_id", "scope_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrScopeNotFound)
