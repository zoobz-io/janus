// Package handlers provides HTTP endpoint handlers for the public API.
package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for API handler responses.
var (
	ErrUserNotFound        = rocco.ErrNotFound.WithMessage("user not found")
	ErrSessionNotFound     = rocco.ErrNotFound.WithMessage("session not found")
	ErrIdentityNotFound    = rocco.ErrNotFound.WithMessage("linked identity not found")
	ErrTenantNotFound      = rocco.ErrNotFound.WithMessage("tenant not found")
	ErrTenantRequired      = rocco.ErrForbidden.WithMessage("tenant membership required — create a tenant via POST /me/tenants")
	ErrInsufficientRole    = rocco.ErrForbidden.WithMessage("admin or owner role required")
	ErrMembershipNotFound  = rocco.ErrNotFound.WithMessage("membership not found")
	ErrApplicationNotFound = rocco.ErrNotFound.WithMessage("application not found")
	ErrAlreadyExists       = rocco.ErrConflict.WithMessage("already exists")
)
