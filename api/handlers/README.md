# handlers

The [rocco](https://github.com/zoobz-io/rocco) endpoints for the public API. Each is an
unexported module-level var built with the generic form — `rocco.GET[Req, Resp]`,
`rocco.PUT[Req, Resp]`, and so on — that resolves its contract from context with
`sum.MustUse`, calls it, and maps the result through a [transformer](../transformers/).
Handlers orchestrate; the business logic lives behind the contract. For the endpoint
table, the request flow, and the authentication model, see
[the four-package pattern](../README.md#the-four-package-pattern) and the
[surface overview](../README.md).

The endpoints are unexported because the two registration functions are the only public
seam. [`All()`](handlers.go) returns the eleven authenticated endpoints — `getMyProfile`,
`updateMyProfile`, `createMyTenant`, `listMySessions`, `revokeMySession`,
`revokeAllMySessions`, `listMyAccounts`, `unlinkMyAccount`, `listApplications`,
`listMyApplications`, `getMyAuthorization`. `AllWithAuth(auth *AuthHandlers)` returns
`All()` plus the OIDC `/auth/login`, `/auth/callback`, and `/auth/logout` handlers, which
are built at startup from session config and so cannot be package-level vars. One detail
worth pinning down against the code: the profile edit is `PUT /me`
(`rocco.PUT[wire.UpdateProfileRequest, wire.UserResponse]`), a full replace — not `PATCH`.
