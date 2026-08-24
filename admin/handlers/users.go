package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
)

var listUsers = rocco.GET[rocco.NoBody, wire.UserListResponse]("/users", func(r *rocco.Request[rocco.NoBody]) (wire.UserListResponse, error) {
	users := sum.MustUse[contracts.Users](r)

	// ?email= narrows to a single user lookup.
	if email := r.Params.Query["email"]; email != "" {
		user, err := users.GetUserByEmail(r, email)
		if err != nil {
			return wire.UserListResponse{Users: []wire.UserResponse{}}, nil
		}
		return wire.UserListResponse{Users: []wire.UserResponse{transformers.UserToResponse(user)}, Total: 1}, nil
	}

	result, err := users.List(r, offsetPage(r.Params))
	if err != nil {
		return wire.UserListResponse{}, err
	}
	return wire.UserListResponse{Users: transformers.UsersToResponse(result.Items), Total: result.Total}, nil
}).
	WithName("list-users").
	WithSummary("List users").
	WithScopes("directory:read").
	WithDescription("Returns a paginated list of users. Pass email to look up a single user by address.").
	WithTags("Users").
	WithQueryParams("email", "limit", "offset").
	WithAuthentication()

var searchUsers = rocco.POST[wire.SearchUsersRequest, wire.UserSearchResponse]("/users/search", func(r *rocco.Request[wire.SearchUsersRequest]) (wire.UserSearchResponse, error) {
	users := sum.MustUse[contracts.Users](r)

	params, number, size := transformers.ResolveUserSearch(r.Body)
	result, err := users.Search(r, params)
	if err != nil {
		return wire.UserSearchResponse{}, err
	}
	return transformers.UserSearchToResponse(result, number, size), nil
}).
	WithName("search-users").
	WithSummary("Search users").
	WithScopes("directory:read").
	WithDescription("Query users with text search over email and display name, status faceting, date-range filters, sorting, and pagination. An empty body returns page 1, size 25, sorted updated_at desc, unfiltered.").
	WithTags("Users").
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var getUser = rocco.GET[rocco.NoBody, wire.UserResponse]("/users/{user_id}", func(r *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUser(r, pathID(r.Params, "user_id"))
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}
	return transformers.UserToResponse(user), nil
}).
	WithName("get-user").
	WithSummary("Get a user").
	WithScopes("directory:read").
	WithDescription("Fetches a single user by ID.").
	WithTags("Users").
	WithPathParams("user_id").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var createUser = rocco.POST[wire.CreateUserRequest, wire.UserResponse]("/users", func(r *rocco.Request[wire.CreateUserRequest]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](r)
	user, err := users.CreateUser(r, r.Body.Email, r.Body.DisplayName)
	if err != nil {
		return wire.UserResponse{}, err
	}
	return transformers.UserToResponse(user), nil
}).
	WithName("create-user").
	WithSummary("Create a user").
	WithScopes("users:manage").
	WithDescription("Creates a user directly. Normally users are provisioned automatically via the OIDC login flow.").
	WithTags("Users").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateUser = rocco.PATCH[wire.UpdateUserRequest, wire.UserResponse]("/users/{user_id}", func(r *rocco.Request[wire.UpdateUserRequest]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](r)
	user, err := users.Update(r, pathID(r.Params, "user_id"), r.Body.DisplayName, r.Body.Status)
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}
	return transformers.UserToResponse(user), nil
}).
	WithName("update-user").
	WithSummary("Update a user").
	WithScopes("users:manage").
	WithDescription("Updates a user's display name and status. Set status to 'inactive' to disable the account.").
	WithTags("Users").
	WithPathParams("user_id").
	WithAuthentication().
	WithErrors(ErrUserNotFound, rocco.ErrValidationFailed)
