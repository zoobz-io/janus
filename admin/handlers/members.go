package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/models"
)

var listMembers = rocco.GET[rocco.NoBody, wire.MemberListResponse]("/tenants/{tenant_id}/members", func(r *rocco.Request[rocco.NoBody]) (wire.MemberListResponse, error) {
	members := sum.MustUse[contracts.Memberships](r)
	result, err := members.ListByTenant(r, pathID(r.Params, "tenant_id"), offsetPage(r.Params))
	if err != nil {
		return wire.MemberListResponse{}, err
	}
	return wire.MemberListResponse{Members: transformers.MembersToResponse(result.Items), Total: result.Total}, nil
}).
	WithSummary("List tenant members").
	WithDescription("Returns the members of a tenant with their roles, paginated.").
	WithTags("Tenants").
	WithPathParams("tenant_id").
	WithQueryParams("limit", "offset").
	WithAuthentication()

var addMember = rocco.POST[wire.AddMemberRequest, wire.MemberResponse]("/tenants/{tenant_id}/members", func(r *rocco.Request[wire.AddMemberRequest]) (wire.MemberResponse, error) {
	members := sum.MustUse[contracts.Memberships](r)
	mem, err := members.Create(r, r.Body.UserID, pathID(r.Params, "tenant_id"), r.Body.Role)
	if err != nil {
		return wire.MemberResponse{}, err
	}
	return transformers.MemberToResponse(mem), nil
}).
	WithSummary("Add a tenant member").
	WithDescription("Adds a user to a tenant with a role (viewer, editor, admin, or owner).").
	WithTags("Tenants").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateMemberRole = rocco.PATCH[wire.UpdateMemberRoleRequest, wire.MemberResponse]("/tenants/{tenant_id}/members/{user_id}", func(r *rocco.Request[wire.UpdateMemberRoleRequest]) (wire.MemberResponse, error) {
	members := sum.MustUse[contracts.Memberships](r)
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")

	current, err := members.GetByUserAndTenant(r, userID, tenantID)
	if err != nil {
		return wire.MemberResponse{}, err
	}
	if current == nil {
		return wire.MemberResponse{}, ErrMembershipNotFound
	}
	// Demoting the last owner is not allowed.
	if current.Role == models.UserRoleOwner && r.Body.Role != models.UserRoleOwner {
		if ownerErr := authz.RequireOwnerExists(r, members, tenantID, userID); ownerErr != nil {
			return wire.MemberResponse{}, ErrLastOwner
		}
	}

	mem, err := members.UpdateRole(r, userID, tenantID, r.Body.Role)
	if err != nil {
		return wire.MemberResponse{}, ErrMembershipNotFound
	}
	return transformers.MemberToResponse(mem), nil
}).
	WithSummary("Update a member's role").
	WithDescription("Changes a member's role within a tenant. Demoting the last owner is rejected.").
	WithTags("Tenants").
	WithPathParams("tenant_id", "user_id").
	WithAuthentication().
	WithErrors(ErrMembershipNotFound, ErrLastOwner, rocco.ErrValidationFailed)

var removeMember = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tenants/{tenant_id}/members/{user_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	members := sum.MustUse[contracts.Memberships](r)
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")

	current, err := members.GetByUserAndTenant(r, userID, tenantID)
	if err != nil {
		return rocco.NoBody{}, err
	}
	if current == nil {
		return rocco.NoBody{}, ErrMembershipNotFound
	}
	// Removing the last owner is not allowed.
	if current.Role == models.UserRoleOwner {
		if err := authz.RequireOwnerExists(r, members, tenantID, userID); err != nil {
			return rocco.NoBody{}, ErrLastOwner
		}
	}

	if err := members.Remove(r, userID, tenantID); err != nil {
		return rocco.NoBody{}, ErrMembershipNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Remove a tenant member").
	WithDescription("Removes a user from a tenant. Removing the last owner is rejected.").
	WithTags("Tenants").
	WithPathParams("tenant_id", "user_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrMembershipNotFound, ErrLastOwner)
