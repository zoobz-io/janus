package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/database/models"
)

const (
	// adminAppSlug is the application the admin API authorizes against — janus's
	// own admin portal, licensed and granted like any other app.
	adminAppSlug = "janus-admin"
	// adminIssuer is the session issuer for admin-minted bearer sessions. A
	// bearer credential is only valid at the admin surface if it was issued for
	// it; sessions minted over the mesh by other apps carry that app's CN.
	adminIssuer = "janus-admin"
)

// sessionValidator is the session-lookup subset the admin authenticator needs.
type sessionValidator interface {
	ValidateByToken(ctx context.Context, token string) (*models.Session, error)
}

// UserLookup is the user-lookup subset the admin authenticator needs.
type UserLookup interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
}

// entitlementResolver is the authz subset the admin authenticator needs.
type entitlementResolver interface {
	ForApplication(ctx context.Context, userID, appSlug string) (*models.Application, []models.AuthorizedTenant, error)
}

// NewAdminAuthenticator builds a rocco authenticator for the admin API that
// enforces operator entitlement, not mere session presence. Two credential
// paths, both funneled through the same entitlement gate:
//
//   - cookie: a first-party login session (issued_by "janus"), forwarded from
//     the admin UI. The extractor yields the user; entitlement is then resolved.
//   - bearer: must be an admin-issued session (issued_by "janus-admin"). A
//     session minted by any other mesh app is that app's credential and is
//     refused, regardless of who it belongs to.
//
// A user with zero janus-admin-entitled tenants is rejected at the door.
func NewAdminAuthenticator(
	sessions sessionValidator,
	users UserLookup,
	entitlements entitlementResolver,
	cookieExtractor func(context.Context, *http.Request) (rocco.Identity, error),
) func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(ctx context.Context, r *http.Request) (rocco.Identity, error) {
		userID, email, err := resolveAdminCredential(ctx, r, sessions, users, cookieExtractor)
		if err != nil {
			return nil, err
		}

		_, tenants, err := entitlements.ForApplication(ctx, userID, adminAppSlug)
		if err != nil {
			return nil, fmt.Errorf("resolving admin entitlement: %w", err)
		}
		if len(tenants) == 0 {
			return nil, fmt.Errorf("no janus-admin entitlement for user %s", userID)
		}

		return NewAdminIdentity(userID, email, tenants), nil
	}
}

// resolveAdminCredential authenticates the request to a user, enforcing the
// per-surface issuer rule. It does not check entitlement — that is the caller's
// gate.
func resolveAdminCredential(
	ctx context.Context,
	r *http.Request,
	sessions sessionValidator,
	users UserLookup,
	cookieExtractor func(context.Context, *http.Request) (rocco.Identity, error),
) (userID, email string, err error) {
	// Cookie first: a first-party login session. Cookie sessions are only ever
	// minted by the login flow (issued_by "janus") and the cookie is signed, so
	// the path is structurally first-party — no separate issuer check needed.
	if cookieExtractor != nil {
		if identity, cerr := cookieExtractor(ctx, r); cerr == nil && identity != nil {
			return identity.ID(), identity.Email(), nil
		}
	}

	// Bearer fallback — must be admin-issued.
	token, terr := extractBearerToken(r)
	if terr != nil {
		return "", "", fmt.Errorf("no valid authentication: missing session cookie and %w", terr)
	}
	sess, serr := sessions.ValidateByToken(ctx, token)
	if serr != nil {
		return "", "", fmt.Errorf("session validation failed: %w", serr)
	}
	if sess == nil {
		return "", "", fmt.Errorf("invalid or expired session token")
	}
	if sess.IssuedBy != adminIssuer {
		return "", "", fmt.Errorf("bearer session issued by %q is not an admin credential", sess.IssuedBy)
	}

	user, uerr := users.GetUser(ctx, sess.UserID)
	if uerr != nil {
		return "", "", fmt.Errorf("user lookup failed: %w", uerr)
	}
	return user.ID, user.Email, nil
}
