package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/rocco/session"

	"github.com/zoobz-io/janus/stores"
)

// NewAuthenticator creates a rocco-compatible authenticator that supports
// two auth methods:
//  1. Cookie-based (OAuth login flow) — tried first
//  2. Bearer token (mesh-issued sessions) — fallback
//
// If cookieExtractor is nil, only bearer tokens are accepted.
func NewAuthenticator(sessions *stores.Sessions, users *stores.Users, cookieExtractor func(context.Context, *http.Request) (rocco.Identity, error)) func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(ctx context.Context, r *http.Request) (rocco.Identity, error) {
		// Try cookie-based auth first (if configured).
		if cookieExtractor != nil {
			identity, err := cookieExtractor(ctx, r)
			if err == nil && identity != nil {
				return identity, nil
			}
		}

		// Fall back to bearer token.
		token, err := extractBearerToken(r)
		if err != nil {
			return nil, fmt.Errorf("no valid authentication: missing session cookie and %w", err)
		}

		sess, err := sessions.ValidateByToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("session validation failed: %w", err)
		}
		if sess == nil {
			return nil, fmt.Errorf("invalid or expired session token")
		}

		user, err := users.GetUser(ctx, sess.UserID)
		if err != nil {
			return nil, fmt.Errorf("user lookup failed: %w", err)
		}

		return &JanusIdentity{
			userID: user.ID,
			email:  user.Email,
		}, nil
	}
}

// CookieExtractor returns a rocco-compatible authenticator function that
// validates session cookies. Wraps rocco/session.Extractor.
func CookieExtractor(store session.Store, cookieCfg session.CookieConfig) func(context.Context, *http.Request) (rocco.Identity, error) {
	return session.Extractor(store, cookieCfg)
}

// extractBearerToken pulls the token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}
