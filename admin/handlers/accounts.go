package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listUserAccounts = rocco.GET[rocco.NoBody, wire.AccountListResponse]("/users/{user_id}/accounts", func(r *rocco.Request[rocco.NoBody]) (wire.AccountListResponse, error) {
	accounts := sum.MustUse[contracts.Accounts](r)
	list, err := accounts.ListByUser(r, pathID(r.Params, "user_id"))
	if err != nil {
		return wire.AccountListResponse{}, err
	}
	return wire.AccountListResponse{Accounts: transformers.AccountsToResponse(list)}, nil
}).
	WithName("list-user-accounts").
	WithSummary("List a user's linked accounts").
	WithDescription("Returns the external identity-provider accounts linked to a user.").
	WithTags("Users").
	WithPathParams("user_id").
	WithAuthentication()

var unlinkUserAccount = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/users/{user_id}/accounts/{account_id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	accounts := sum.MustUse[contracts.Accounts](r)
	if err := accounts.Unlink(r, pathID(r.Params, "account_id"), pathID(r.Params, "user_id")); err != nil {
		return rocco.NoBody{}, ErrAccountNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithName("unlink-user-account").
	WithSummary("Unlink a user's account").
	WithDescription("Removes a linked external account from a user.").
	WithTags("Users").
	WithPathParams("user_id", "account_id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrAccountNotFound)
