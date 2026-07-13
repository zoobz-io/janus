package transformers

import (
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/models"
)

// AccountToResponse transforms a Account model to an API response.
func AccountToResponse(l *models.Account) wire.AccountResponse {
	return wire.AccountResponse{
		ID:              l.ID,
		Provider:        l.Provider,
		ExternalSubject: l.ExternalSubject,
		LinkedAt:        l.LinkedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// AccountsToResponse transforms a slice of Account models to responses.
func AccountsToResponse(identities []*models.Account) []wire.AccountResponse {
	result := make([]wire.AccountResponse, len(identities))
	for i, l := range identities {
		result[i] = AccountToResponse(l)
	}
	return result
}
