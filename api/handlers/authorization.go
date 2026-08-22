package handlers

import (
	"errors"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/internal/authz"
)

// Self-service — what am I entitled to in an application? This is the
// endpoint a consumer application's auth provider dials after login to turn
// a janus session into that app's own user: tenants, roles, and effective
// scopes, resolved from License + Grant (+ tier features).

var getMyAuthorization = rocco.GET[rocco.NoBody, wire.AuthorizationResponse]("/me/authorization/{app_slug}", func(r *rocco.Request[rocco.NoBody]) (wire.AuthorizationResponse, error) {
	slug := pathID(r.Params, "app_slug")
	authorizations := sum.MustUse[contracts.Authorizations](r)

	app, tenants, err := authorizations.ForApplication(r, r.Identity.ID(), slug)
	if err != nil {
		if errors.Is(err, authz.ErrApplicationNotFound) {
			return wire.AuthorizationResponse{}, ErrApplicationNotFound
		}
		return wire.AuthorizationResponse{}, err
	}

	return transformers.AuthorizationToResponse(r.Identity.ID(), app, tenants), nil
}).
	WithName("get-my-authorization").
	WithSummary("Get my authorization for an application").
	WithTags("Applications").
	WithPathParams("app_slug").
	WithAuthentication().
	WithErrors(ErrApplicationNotFound)
