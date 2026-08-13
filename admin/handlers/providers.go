package handlers

import (
	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

var listProviders = rocco.GET[rocco.NoBody, wire.ProviderListResponse]("/providers", func(_ *rocco.Request[rocco.NoBody]) (wire.ProviderListResponse, error) {
	return wire.ProviderListResponse{Providers: []string{
		models.ProviderZitadel,
		models.ProviderAuth0,
		models.ProviderGitHub,
		models.ProviderGoogle,
	}}, nil
}).
	WithSummary("List supported identity providers").
	WithDescription("Returns the external identity providers janus supports for account linking.").
	WithTags("Providers").
	WithAuthentication()
