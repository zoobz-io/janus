package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listTenantMembers = rocco.GET[rocco.NoBody, wire.MemberListResponse]("/tenants/{tenant_id}/members", func(r *rocco.Request[rocco.NoBody]) (wire.MemberListResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleViewer, models.UserRoleEditor, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.MemberListResponse{}, err
	}

	memberships := sum.MustUse[contracts.Memberships](r)
	mems, err := memberships.ListByTenant(r, tenantID, models.OffsetPage{Limit: 100})
	if err != nil {
		return wire.MemberListResponse{}, err
	}
	return wire.MemberListResponse{
		Members: transformers.MembersToResponse(mems.Items),
	}, nil
}).
	WithSummary("List tenant members").
	WithTags("Members").
	WithPathParams("tenant_id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound)

var addTenantMember = rocco.POST[wire.AddMemberRequest, wire.MemberResponse]("/tenants/{tenant_id}/members", func(r *rocco.Request[wire.AddMemberRequest]) (wire.MemberResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.MemberResponse{}, err
	}

	memberships := sum.MustUse[contracts.Memberships](r)

	// Check if already a member.
	existing, err := memberships.GetByUserAndTenant(r, r.Body.UserID, tenantID)
	if err != nil {
		return wire.MemberResponse{}, err
	}
	if existing != nil {
		return wire.MemberResponse{}, ErrAlreadyExists
	}

	mem, err := memberships.Create(r, r.Body.UserID, tenantID, r.Body.Role)
	if err != nil {
		return wire.MemberResponse{}, err
	}
	return transformers.MemberToResponse(mem), nil
}).
	WithSummary("Add a member to a tenant").
	WithTags("Members").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, ErrAlreadyExists, rocco.ErrValidationFailed)

var updateTenantMemberRole = rocco.PUT[wire.UpdateMemberRoleRequest, wire.MemberResponse]("/tenants/{tenant_id}/members/{user_id}", func(r *rocco.Request[wire.UpdateMemberRoleRequest]) (wire.MemberResponse, error) {
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return wire.MemberResponse{}, err
	}

	memberships := sum.MustUse[contracts.Memberships](r)
	mem, lookupErr := memberships.GetByUserAndTenant(r, userID, tenantID)
	if lookupErr != nil || mem == nil {
		return wire.MemberResponse{}, ErrMembershipNotFound
	}

	if deleteErr := memberships.Delete(r, mem.ID); deleteErr != nil {
		return wire.MemberResponse{}, deleteErr
	}
	updated, createErr := memberships.Create(r, userID, tenantID, r.Body.Role)
	if createErr != nil {
		return wire.MemberResponse{}, createErr
	}
	return transformers.MemberToResponse(updated), nil
}).
	WithSummary("Update a member's role").
	WithTags("Members").
	WithPathParams("tenant_id", "user_id").
	WithAuthentication().
	WithErrors(ErrInsufficientRole, ErrMembershipNotFound, rocco.ErrValidationFailed)

var removeTenantMember = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tenants/{tenant_id}/members/{user_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tenantID := pathID(r.Params, "tenant_id")
	userID := pathID(r.Params, "user_id")
	if _, err := requireRole(r, r.Identity.ID(), tenantID, models.UserRoleAdmin, models.UserRoleOwner); err != nil {
		return rocco.NoBody{}, err
	}

	memberships := sum.MustUse[contracts.Memberships](r)
	mem, err := memberships.GetByUserAndTenant(r, userID, tenantID)
	if err != nil || mem == nil {
		return rocco.NoBody{}, ErrMembershipNotFound
	}
	if err := memberships.Delete(r, mem.ID); err != nil {
		return rocco.NoBody{}, err
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Remove a member from a tenant").
	WithTags("Members").
	WithPathParams("tenant_id", "user_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrInsufficientRole, ErrMembershipNotFound)
