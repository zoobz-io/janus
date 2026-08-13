package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var createMyTenant = rocco.POST[wire.CreateTenantRequest, wire.TenantResponse]("/me/tenants", func(r *rocco.Request[wire.CreateTenantRequest]) (wire.TenantResponse, error) {
	tenants := sum.MustUse[contracts.Tenants](r)
	memberships := sum.MustUse[contracts.Memberships](r)

	tenant, err := tenants.CreateTenant(r, r.Body.Name, r.Body.Slug)
	if err != nil {
		return wire.TenantResponse{}, err
	}

	if _, err := memberships.Create(r, r.Identity.ID(), tenant.ID, models.UserRoleOwner); err != nil {
		return wire.TenantResponse{}, err
	}

	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Create a new tenant").
	WithTags("Tenants").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)
