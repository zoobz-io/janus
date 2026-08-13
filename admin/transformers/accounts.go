package transformers

import (
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

// AccountToResponse transforms an Account model to an admin API response.
func AccountToResponse(a *models.Account) wire.AccountResponse {
	return wire.AccountResponse{
		ID:              a.ID,
		Provider:        a.Provider,
		ExternalSubject: a.ExternalSubject,
		LinkedAt:        a.LinkedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// AccountsToResponse transforms a slice of Account models to responses.
func AccountsToResponse(accounts []*models.Account) []wire.AccountResponse {
	result := make([]wire.AccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = AccountToResponse(a)
	}
	return result
}
