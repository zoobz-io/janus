package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var getMyProfile = rocco.GET[rocco.NoBody, wire.UserResponse]("/me", func(r *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
	store := sum.MustUse[contracts.Users](r)
	user, err := store.GetUser(r, r.Identity.ID())
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}
	return transformers.UserToResponse(user), nil
}).
	WithSummary("Get my profile").
	WithTags("Profile").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var updateMyProfile = rocco.PUT[wire.UpdateProfileRequest, wire.UserResponse]("/me", func(r *rocco.Request[wire.UpdateProfileRequest]) (wire.UserResponse, error) {
	store := sum.MustUse[contracts.Users](r)
	user, err := store.UpdateDisplayName(r, r.Identity.ID(), r.Body.DisplayName)
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}
	return transformers.UserToResponse(user), nil
}).
	WithSummary("Update my profile").
	WithTags("Profile").
	WithAuthentication().
	WithErrors(ErrUserNotFound, rocco.ErrValidationFailed)
