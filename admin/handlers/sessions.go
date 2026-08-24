package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listUserSessions = rocco.GET[rocco.NoBody, wire.SessionListResponse]("/users/{user_id}/sessions", func(r *rocco.Request[rocco.NoBody]) (wire.SessionListResponse, error) {
	sessions := sum.MustUse[contracts.Sessions](r)
	list, err := sessions.ListSessionsByUser(r, pathID(r.Params, "user_id"))
	if err != nil {
		return wire.SessionListResponse{}, err
	}
	return wire.SessionListResponse{Sessions: transformers.SessionsToResponse(list)}, nil
}).
	WithName("list-user-sessions").
	WithSummary("List a user's sessions").
	WithScopes("directory:read").
	WithDescription("Returns the active sessions for a user.").
	WithTags("Users").
	WithPathParams("user_id").
	WithAuthentication()

var revokeUserSession = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/users/{user_id}/sessions/{session_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	sessions := sum.MustUse[contracts.Sessions](r)
	if err := sessions.RevokeSession(r, pathID(r.Params, "session_id"), pathID(r.Params, "user_id")); err != nil {
		return rocco.NoBody{}, ErrSessionNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithName("revoke-user-session").
	WithSummary("Revoke a user's session").
	WithScopes("users:manage").
	WithDescription("Revokes a single session by ID.").
	WithTags("Users").
	WithPathParams("user_id", "session_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrSessionNotFound)

var revokeAllUserSessions = rocco.DELETE[rocco.NoBody, wire.RevokeAllSessionsResponse]("/users/{user_id}/sessions", func(r *rocco.Request[rocco.NoBody]) (wire.RevokeAllSessionsResponse, error) {
	sessions := sum.MustUse[contracts.Sessions](r)
	count, err := sessions.RevokeUserSessions(r, pathID(r.Params, "user_id"))
	if err != nil {
		return wire.RevokeAllSessionsResponse{}, err
	}
	return wire.RevokeAllSessionsResponse{RevokedCount: count}, nil
}).
	WithName("revoke-all-user-sessions").
	WithSummary("Revoke all of a user's sessions").
	WithScopes("users:manage").
	WithDescription("Revokes every active session for a user and returns the count revoked.").
	WithTags("Users").
	WithPathParams("user_id").
	WithAuthentication()
