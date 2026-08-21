package handlers

import (
	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/transformers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var createMyTenant = rocco.POST[wire.CreateTenantRequest, wire.TenantResponse]("/me/tenants", func(r *rocco.Request[wire.CreateTenantRequest]) (wire.TenantResponse, error) {
	provisioning := sum.MustUse[contracts.Provisioning](r)

	tenant, _, err := provisioning.CreateTenantWithOwner(r, r.Body.Name, r.Body.Slug, r.Identity.ID())
	if err != nil {
		return wire.TenantResponse{}, err
	}

	return transformers.TenantToResponse(tenant), nil
}).
	WithSummary("Create a new tenant").
	WithTags("Tenants").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)
