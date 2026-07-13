// Package handlers provides HTTP endpoint handlers for the admin API.
package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for admin API handler responses.
var (
	ErrApplicationNotFound = rocco.ErrNotFound.WithMessage("application not found")
	ErrTenantNotFound      = rocco.ErrNotFound.WithMessage("tenant not found")
	ErrUserNotFound        = rocco.ErrNotFound.WithMessage("user not found")
	ErrMembershipNotFound  = rocco.ErrNotFound.WithMessage("membership not found")
	ErrSessionNotFound     = rocco.ErrNotFound.WithMessage("session not found")
	ErrAccountNotFound     = rocco.ErrNotFound.WithMessage("account not found")
	ErrScopeNotFound       = rocco.ErrNotFound.WithMessage("scope not found")
	ErrTierNotFound        = rocco.ErrNotFound.WithMessage("tier not found")
	ErrLastOwner           = rocco.ErrForbidden.WithMessage("cannot remove or demote the last owner")
)
