package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listMyIdentities = rocco.GET[rocco.NoBody, wire.LinkedIdentityListResponse]("/me/identities", func(r *rocco.Request[rocco.NoBody]) (wire.LinkedIdentityListResponse, error) {
	store := sum.MustUse[contracts.LinkedIdentities](r)
	identities, err := store.ListByUser(r, r.Identity.ID())
	if err != nil {
		return wire.LinkedIdentityListResponse{}, err
	}
	return wire.LinkedIdentityListResponse{
		Identities: transformers.LinkedIdentitiesToResponse(identities),
	}, nil
}).
	WithSummary("List my linked identities").
	WithTags("Identities").
	WithAuthentication()

var unlinkMyIdentity = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/me/identities/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.LinkedIdentities](r)
	if err := store.Unlink(r, id, r.Identity.ID()); err != nil {
		return rocco.NoBody{}, ErrIdentityNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Unlink an identity provider").
	WithTags("Identities").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrIdentityNotFound)
