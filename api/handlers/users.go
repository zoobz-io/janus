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

var getMyProfile = rocco.GET[rocco.NoBody, wire.UserResponse]("/me", func(r *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](r)
	memberships := sum.MustUse[contracts.Memberships](r)
	tenants := sum.MustUse[contracts.Tenants](r)

	// A store failure must surface as an internal error, not a 404 —
	// masking outages as not-found misleads both clients and operators.
	user, err := users.GetUser(r, r.Identity.ID())
	if errors.Is(err, stores.ErrNotFound) {
		return wire.UserResponse{}, ErrUserNotFound
	}
	if err != nil {
		return wire.UserResponse{}, err
	}

	mems, err := memberships.ListByUser(r, user.ID)
	if err != nil {
		return wire.UserResponse{}, err
	}

	tenantNames := make(map[string]string, len(mems))
	for _, m := range mems {
		t, err := tenants.GetTenant(r, m.TenantID)
		if err == nil {
			tenantNames[m.TenantID] = t.Name
		}
	}

	return transformers.UserToResponse(user, mems, tenantNames), nil
}).
	WithName("get-my-profile").
	WithSummary("Get my profile").
	WithTags("Profile").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var updateMyProfile = rocco.PUT[wire.UpdateProfileRequest, wire.UserResponse]("/me", func(r *rocco.Request[wire.UpdateProfileRequest]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](r)
	memberships := sum.MustUse[contracts.Memberships](r)
	tenants := sum.MustUse[contracts.Tenants](r)

	user, err := users.UpdateDisplayName(r, r.Identity.ID(), r.Body.DisplayName)
	if errors.Is(err, stores.ErrNotFound) {
		return wire.UserResponse{}, ErrUserNotFound
	}
	if err != nil {
		return wire.UserResponse{}, err
	}

	mems, err := memberships.ListByUser(r, user.ID)
	if err != nil {
		return wire.UserResponse{}, err
	}

	tenantNames := make(map[string]string, len(mems))
	for _, m := range mems {
		t, err := tenants.GetTenant(r, m.TenantID)
		if err == nil {
			tenantNames[m.TenantID] = t.Name
		}
	}

	return transformers.UserToResponse(user, mems, tenantNames), nil
}).
	WithName("update-my-profile").
	WithSummary("Update my profile").
	WithTags("Profile").
	WithAuthentication().
	WithErrors(ErrUserNotFound, rocco.ErrValidationFailed)
