# contracts

The Go interfaces the public API is allowed to call. Handlers depend on these
interfaces; [`database/stores`](../../database/stores/) implements them. The seam is the
point: the *same* store instance backs both this narrow public contract and the broad
admin one — each surface simply declares a different method set. See
[the four-package pattern](../README.md#the-four-package-pattern) for where contracts sit
in the request flow.

What makes the public contracts *public* is how little they expose. `Users` is the whole
argument:

```go
type Users interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	UpdateDisplayName(ctx context.Context, id, displayName string) (*models.User, error)
}
```

Read yourself, rename yourself — nothing more. `Sessions`, `Accounts`, `Tenants`,
`Grants`, and `Authorizations` are cut to the same shape: every method is scoped to the
caller's own identity, and there is no cross-tenant or administrative verb anywhere in the
package. `Provisioning` (in [`tenants.go`](tenants.go)) is the one exception to
one-method-does-one-thing: `CreateTenantWithOwner` is a transactional multi-store flow,
satisfied structurally by the store aggregate so the tenant and its owner membership land
atomically — it must not be recomposed from single-store calls.

Two interfaces are defined but drive no public endpoint. [`licenses.go`](licenses.go)
carries `Licenses`, and [`memberships.go`](memberships.go)'s `Memberships` exposes list
and write verbs the public [`All()`](../handlers/handlers.go) never registers — the public
surface only reads memberships, as part of assembling a profile. They live here so both
this surface and the admin surface can register and reuse the same contract types over the
same stores; the public handler set just declines to wire them to routes.

Models are in [`database/models`](../../database/models/); the module is
`github.com/zoobz-io/janus`.
