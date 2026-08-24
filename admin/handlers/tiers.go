package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listTiers = rocco.GET[rocco.NoBody, wire.TierListResponse]("/applications/{app_id}/tiers", func(r *rocco.Request[rocco.NoBody]) (wire.TierListResponse, error) {
	tiers := sum.MustUse[contracts.Tiers](r)
	labels := sum.MustUse[contracts.ApplicationLabels](r)
	list, err := tiers.ListByApplication(r, pathID(r.Params, "app_id"))
	if err != nil {
		return wire.TierListResponse{}, err
	}
	responses, err := transformers.TiersToResponse(r, list, labels)
	if err != nil {
		return wire.TierListResponse{}, err
	}
	return wire.TierListResponse{Tiers: responses}, nil
}).
	WithName("list-tiers").
	WithSummary("List application tiers").
	WithScopes("directory:read").
	WithDescription("Returns the subscription tiers defined for an application, ordered by rank.").
	WithTags("Tiers").
	WithPathParams("app_id").
	WithAuthentication()

var createTier = rocco.POST[wire.CreateTierRequest, wire.TierResponse]("/applications/{app_id}/tiers", func(r *rocco.Request[wire.CreateTierRequest]) (wire.TierResponse, error) {
	tiers := sum.MustUse[contracts.Tiers](r)
	labels := sum.MustUse[contracts.ApplicationLabels](r)
	t, err := tiers.Define(r, pathID(r.Params, "app_id"), r.Body.Slug, r.Body.Name, r.Body.Rank)
	if err != nil {
		return wire.TierResponse{}, err
	}
	return transformers.TierToResponse(r, t, labels)
}).
	WithName("create-tier").
	WithSummary("Define an application tier").
	WithScopes("applications:manage").
	WithDescription("Adds a subscription tier to an application. Bundle scopes into it via its features.").
	WithTags("Tiers").
	WithPathParams("app_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var deleteTier = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/applications/{app_id}/tiers/{tier_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tiers := sum.MustUse[contracts.Tiers](r)
	if err := tiers.Delete(r, pathID(r.Params, "tier_id")); err != nil {
		return rocco.NoBody{}, ErrTierNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithName("delete-tier").
	WithSummary("Delete an application tier").
	WithScopes("applications:manage").
	WithDescription("Removes a tier. Its feature bundles are cascaded; grants that referenced it keep their explicit scopes.").
	WithTags("Tiers").
	WithPathParams("app_id", "tier_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrTierNotFound)
