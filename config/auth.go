package config

import (
	"encoding/hex"
	"errors"
)

// Auth holds OAuth and session cookie configuration.
type Auth struct {
	Issuer            string `env:"APP_AUTH_ISSUER" default:"http://localhost:8085"`
	ClientID          string `env:"APP_AUTH_CLIENT_ID"`
	ClientSecret      string `env:"APP_AUTH_CLIENT_SECRET" secret:"app/auth-client-secret"`
	RedirectURI       string `env:"APP_AUTH_REDIRECT_URI" default:"http://localhost:8080/auth/callback"`
	CookieSignKey     string `env:"APP_AUTH_COOKIE_KEY" secret:"app/auth-cookie-key"`
	PostLoginRedirect string `env:"APP_AUTH_POST_LOGIN_REDIRECT" default:"/me"`
}

// CookieKey returns the cookie signing key as bytes.
func (c Auth) CookieKey() ([]byte, error) {
	return hex.DecodeString(c.CookieSignKey)
}

// Validate checks that the configuration is valid.
func (c Auth) Validate() error {
	if c.Issuer == "" {
		return errors.New("auth issuer URL is required")
	}
	if c.ClientID == "" {
		return errors.New("auth client ID is required")
	}
	if c.RedirectURI == "" {
		return errors.New("auth redirect URI is required")
	}
	if c.CookieSignKey != "" && len(c.CookieSignKey) != 64 {
		return errors.New("auth cookie sign key must be 64 hex characters (32 bytes)")
	}
	return nil
}
