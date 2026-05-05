// Package handlers provides HTTP endpoint handlers for the public API.
package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for API handler responses.
var (
	ErrUserNotFound     = rocco.ErrNotFound.WithMessage("user not found")
	ErrSessionNotFound  = rocco.ErrNotFound.WithMessage("session not found")
	ErrIdentityNotFound = rocco.ErrNotFound.WithMessage("linked identity not found")
)
