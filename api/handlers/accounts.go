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

var listMyAccounts = rocco.GET[rocco.NoBody, wire.AccountListResponse]("/me/accounts", func(r *rocco.Request[rocco.NoBody]) (wire.AccountListResponse, error) {
	store := sum.MustUse[contracts.Accounts](r)
	identities, err := store.ListByUser(r, r.Identity.ID())
	if err != nil {
		return wire.AccountListResponse{}, err
	}
	return wire.AccountListResponse{
		Accounts: transformers.AccountsToResponse(identities),
	}, nil
}).
	WithSummary("List my accounts").
	WithTags("Accounts").
	WithAuthentication()

var unlinkMyAccount = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/me/accounts/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Accounts](r)
	if err := store.Unlink(r, id, r.Identity.ID()); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			return rocco.NoBody{}, ErrAccountNotFound
		}
		return rocco.NoBody{}, err
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Unlink an account").
	WithTags("Accounts").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrAccountNotFound)
