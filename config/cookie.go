package config

import (
	"encoding/hex"
	"errors"
)

// Cookie holds the session cookie signing key shared by every HTTP surface
// that accepts browser sessions. It is the subset of Auth a service loads when
// it validates sessions but never initiates OIDC login (e.g. the admin API),
// so those services don't require the OAuth client configuration.
type Cookie struct {
	SignKey string `env:"APP_AUTH_COOKIE_KEY" secret:"app/auth-cookie-key"`
}

// Key returns the cookie signing key as bytes.
func (c Cookie) Key() ([]byte, error) {
	return hex.DecodeString(c.SignKey)
}

// Validate checks that the configuration is valid.
func (c Cookie) Validate() error {
	if len(c.SignKey) != 64 {
		return errors.New("auth cookie sign key must be 64 hex characters (32 bytes)")
	}
	return nil
}
