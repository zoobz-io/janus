package handlers

import (
	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/models"
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
	WithTags("Providers").
	WithAuthentication()
