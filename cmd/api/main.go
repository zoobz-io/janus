// Package main is the entry point for the janus public HTTP API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco/oauth"
	"github.com/zoobz-io/rocco/session"
	"github.com/zoobz-io/sum"

	apicontracts "github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/handlers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/auth"
	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/internal/observe"
	"github.com/zoobz-io/janus/stores"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting janus api...")
	ctx := context.Background()

	rt, err := boot.Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime: %w", err)
	}
	defer func() { _ = rt.DB.Close() }()
	defer func() { _ = rt.Redis.Close() }()

	if cfgErr := sum.Config[config.App](ctx, rt.K, nil); cfgErr != nil {
		return fmt.Errorf("failed to load app config: %w", cfgErr)
	}
	if cfgErr := sum.Config[config.Auth](ctx, rt.K, nil); cfgErr != nil {
		return fmt.Errorf("failed to load auth config: %w", cfgErr)
	}

	// Public API contracts.
	sum.Register[apicontracts.Users](rt.K, rt.Stores.Users)
	sum.Register[apicontracts.Sessions](rt.K, rt.Stores.Sessions)
	sum.Register[apicontracts.Accounts](rt.K, rt.Stores.Accounts)
	sum.Register[apicontracts.Memberships](rt.K, rt.Stores.Memberships)
	sum.Register[apicontracts.Tenants](rt.K, rt.Stores.Tenants)
	sum.Register[apicontracts.Applications](rt.K, rt.Stores.Applications)
	sum.Register[apicontracts.Licenses](rt.K, rt.Stores.Licenses)
	sum.Register[apicontracts.Grants](rt.K, rt.Stores.Grants)
	sum.Register[apicontracts.Authorizations](rt.K, authz.NewEntitlements(
		rt.Stores.Applications, rt.Stores.Licenses, rt.Stores.Grants,
		rt.Stores.Memberships, rt.Stores.Tenants, rt.Stores.Features, rt.Stores.Scopes,
	))
	// Internal services (not exposed on the HTTP API).
	sum.Register[apicontracts.Config](rt.K, rt.Stores.Config)

	wire.RegisterBoundaries(rt.K)
	sum.Freeze(rt.K)
	capitan.Emit(ctx, events.StartupServicesReady)

	// Observability.
	otelProviders, err := boot.OTEL(ctx, "janus-api")
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()

	if _, schemaErr := observe.StartSchemaSync(ctx, ap); schemaErr != nil {
		return fmt.Errorf("failed to start aperture schema sync: %w", schemaErr)
	}

	// Authentication (cookie + bearer).
	authCfg := sum.MustUse[config.Auth](ctx)
	sessionStore := auth.NewSessionStore(rt.Stores.Sessions, rt.Stores.Users, rt.Redis)

	cookieKey, err := authCfg.CookieKey()
	if err != nil {
		return fmt.Errorf("failed to decode cookie sign key: %w", err)
	}
	cookieCfg := session.CookieConfig{SignKey: cookieKey}
	cookieExtractor := auth.CookieExtractor(sessionStore, cookieCfg)
	authenticator := auth.NewAuthenticator(rt.Stores.Sessions, rt.Stores.Users, cookieExtractor)
	rt.Svc.Engine().WithAuthenticator(authenticator)

	// OAuth login/callback/logout handlers. The issuer's endpoints come from
	// OIDC discovery — never assumed to live under the issuer prefix (Google
	// spreads them across hosts).
	endpoints, err := auth.Discover(ctx, authCfg.Issuer)
	if err != nil {
		return fmt.Errorf("failed to discover OIDC endpoints for %s: %w", authCfg.Issuer, err)
	}

	oauthCfg := oauth.Config{
		Name:         authCfg.Provider,
		AuthURL:      endpoints.AuthURL,
		TokenURL:     endpoints.TokenURL,
		ClientID:     authCfg.ClientID,
		ClientSecret: authCfg.ClientSecret,
		RedirectURI:  authCfg.RedirectURI,
		Scopes:       []string{"openid", "profile", "email"},
	}
	sessionCfg := session.Config{
		OAuth:       oauthCfg,
		Store:       sessionStore,
		Cookie:      cookieCfg,
		RedirectURL: authCfg.PostLoginRedirect,
		Resolve:     resolveOAuth(authCfg.Provider, endpoints.UserinfoURL, rt.Stores),
	}
	authHandlers, err := handlers.NewAuthHandlers(sessionCfg)
	if err != nil {
		return fmt.Errorf("failed to create auth handlers: %w", err)
	}

	rt.Svc.Handle(handlers.AllWithAuth(authHandlers)...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting api server on port %d...", appCfg.Port)
	return rt.Svc.Run("", appCfg.Port)
}

// resolveOAuth returns the Resolve callback for the session config. It maps
// the IdP's userinfo onto a janus user in three steps: a previously linked
// account wins; otherwise a verified email claim links the account to the
// existing user with that address (how seeded/pre-provisioned operators get
// their IdP identity attached on first login); otherwise a brand-new user is
// registered (with no tenant — the user creates one via POST /me/tenants).
//
// Email linking trusts the configured issuer to assert addresses it has
// verified. Acceptable with a single trusted issuer; revisit before
// federating multiple issuers.
func resolveOAuth(provider, userinfoURL string, st *stores.Stores) func(context.Context, *oauth.TokenResponse) (*session.Data, error) {
	return func(ctx context.Context, tokens *oauth.TokenResponse) (*session.Data, error) {
		info, err := fetchUserInfo(ctx, userinfoURL, tokens.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("fetching user info: %w", err)
		}

		// A linked account is the fast path.
		account, err := st.Accounts.GetByProviderSubject(ctx, provider, info.Sub)
		if err != nil {
			return nil, fmt.Errorf("looking up account: %w", err)
		}
		if account != nil {
			user, userErr := st.Users.GetUser(ctx, account.UserID)
			if userErr != nil {
				return nil, fmt.Errorf("looking up user: %w", userErr)
			}
			touchLastSeen(ctx, st, user.ID)
			return &session.Data{UserID: user.ID, Email: user.Email}, nil
		}

		// No account: a verified email attaches this identity to the
		// existing user with that address.
		if info.EmailVerified {
			if user, lookupErr := st.Users.GetUserByEmail(ctx, info.Email); lookupErr == nil && user != nil {
				if _, linkErr := st.Accounts.Link(ctx, user.ID, provider, info.Sub); linkErr != nil {
					return nil, fmt.Errorf("linking identity: %w", linkErr)
				}
				events.IdentityLinked.Emit(ctx, events.IdentityEvent{UserID: user.ID, Provider: provider})
				touchLastSeen(ctx, st, user.ID)
				return &session.Data{UserID: user.ID, Email: user.Email}, nil
			}
		}

		// First contact: register.
		name := info.Name
		if name == "" {
			name = info.Email
		}
		user, err := st.Users.CreateUser(ctx, info.Email, name)
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}
		if _, err := st.Accounts.Link(ctx, user.ID, provider, info.Sub); err != nil {
			return nil, fmt.Errorf("linking identity: %w", err)
		}
		events.UserCreated.Emit(ctx, events.UserEvent{UserID: user.ID, Email: user.Email})
		events.IdentityLinked.Emit(ctx, events.IdentityEvent{UserID: user.ID, Provider: provider})
		return &session.Data{UserID: user.ID, Email: user.Email}, nil
	}
}

// touchLastSeen best-effort stamps activity; failure is logged, not fatal.
func touchLastSeen(ctx context.Context, st *stores.Stores, userID string) {
	if err := st.Users.TouchLastSeen(ctx, userID); err != nil {
		capitan.Warn(ctx, events.LastSeenUpdateFailed, events.OpUserIDKey.Field(userID), events.OpErrorKey.Field(err))
	}
}

// userInfo is the subset of OIDC userinfo claims janus consumes.
type userInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

// fetchUserInfo calls the OIDC userinfo endpoint.
func fetchUserInfo(ctx context.Context, userinfoURL, accessToken string) (*userInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo returned %d: %s", resp.StatusCode, body)
	}

	var info userInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding userinfo: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return nil, fmt.Errorf("userinfo missing sub or email")
	}
	return &info, nil
}
