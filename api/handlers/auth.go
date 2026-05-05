package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/rocco/session"
)

// AuthHandlers holds the login/callback/logout handlers created from session config.
// These are built at startup in main.go since they depend on runtime configuration.
type AuthHandlers struct {
	Login    rocco.Endpoint
	Callback rocco.Endpoint
	Logout   rocco.Endpoint
}

// NewAuthHandlers creates login, callback, and logout handlers from session config.
func NewAuthHandlers(cfg session.Config) (*AuthHandlers, error) {
	login, err := session.NewLoginHandler("/auth/login", cfg)
	if err != nil {
		return nil, err
	}

	callback, err := session.NewCallbackHandler("/auth/callback", cfg)
	if err != nil {
		return nil, err
	}

	logout, err := session.NewLogoutHandler("/auth/logout", cfg, "/")
	if err != nil {
		return nil, err
	}

	return &AuthHandlers{
		Login:    login,
		Callback: callback,
		Logout:   logout,
	}, nil
}
