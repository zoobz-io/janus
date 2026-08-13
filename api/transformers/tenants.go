package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
)

// TenantToResponse transforms a Tenant model to an API response.
func TenantToResponse(t *models.Tenant) wire.TenantResponse {
	return wire.TenantResponse{
		ID:   t.ID,
		Name: t.Name,
		Slug: t.Slug,
	}
}
