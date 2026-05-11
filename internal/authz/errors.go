// Package authz provides portable authorization checks for tenant-scoped
// operations. Usable from both HTTP handlers and gRPC servers.
package authz

import "errors"

var (
	// ErrNotMember indicates the user is not a member of the tenant.
	ErrNotMember = errors.New("not a member of this tenant")
	// ErrInsufficientRole indicates the user's role is not sufficient.
	ErrInsufficientRole = errors.New("insufficient role")
	// ErrLastOwner indicates the operation would remove the last owner.
	ErrLastOwner = errors.New("cannot remove or demote the last owner")
)
