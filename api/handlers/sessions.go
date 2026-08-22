package handlers

import (
	"errors"

	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/stores"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listMySessions = rocco.GET[rocco.NoBody, wire.SessionListResponse]("/me/sessions", func(r *rocco.Request[rocco.NoBody]) (wire.SessionListResponse, error) {
	store := sum.MustUse[contracts.Sessions](r)
	sessions, err := store.ListSessionsByUser(r, r.Identity.ID())
	if err != nil {
		return wire.SessionListResponse{}, err
	}
	return wire.SessionListResponse{
		Sessions: transformers.SessionsToResponse(sessions),
	}, nil
}).
	WithName("list-my-sessions").
	WithSummary("List my sessions").
	WithTags("Sessions").
	WithAuthentication()

var revokeMySession = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/me/sessions/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Sessions](r)
	if err := store.RevokeSession(r, id, r.Identity.ID()); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			return rocco.NoBody{}, ErrSessionNotFound
		}
		return rocco.NoBody{}, err
	}
	return rocco.NoBody{}, nil
}).
	WithName("revoke-my-session").
	WithSummary("Revoke a session").
	WithTags("Sessions").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrSessionNotFound)

var revokeAllMySessions = rocco.DELETE[rocco.NoBody, wire.RevokeAllSessionsResponse]("/me/sessions", func(r *rocco.Request[rocco.NoBody]) (wire.RevokeAllSessionsResponse, error) {
	store := sum.MustUse[contracts.Sessions](r)
	count, err := store.RevokeUserSessions(r, r.Identity.ID())
	if err != nil {
		return wire.RevokeAllSessionsResponse{}, err
	}
	return wire.RevokeAllSessionsResponse{RevokedCount: count}, nil
}).
	WithName("revoke-all-my-sessions").
	WithSummary("Revoke all my sessions").
	WithTags("Sessions").
	WithAuthentication()
