// Package main is the entry point for the janus public HTTP API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"
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
	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/internal/mesh"
	"github.com/zoobz-io/janus/internal/observe"
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

	if err := sum.Config[config.App](ctx, rt.K, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Auth](ctx, rt.K, nil); err != nil {
		return fmt.Errorf("failed to load auth config: %w", err)
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

	// OAuth login/callback/logout handlers.
	identityServer := mesh.NewIdentityServer(
		rt.Stores.Users, rt.Stores.Accounts, rt.Stores.Tenants, rt.Stores.Memberships,
		rt.Stores.Applications, rt.Stores.Licenses, rt.Stores.Grants,
		rt.Stores.Features, rt.Stores.Scopes,
	)

	issuer := strings.TrimSuffix(authCfg.Issuer, "/")
	oauthCfg := oauth.Config{
		Name:         "oidc",
		AuthURL:      issuer + "/oauth/v2/authorize",
		TokenURL:     issuer + "/oauth/v2/token",
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
		Resolve:     resolveOAuth(issuer, identityServer),
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

// resolveOAuth returns the Resolve callback for the session config. It looks up
// the user by account; if not found, creates the user and links the identity
// (but does not create a tenant — the user creates one via POST /me/tenants).
func resolveOAuth(issuer string, identitySvc *mesh.IdentityServer) func(context.Context, *oauth.TokenResponse) (*session.Data, error) {
	userinfoURL := issuer + "/oidc/v1/userinfo"

	return func(ctx context.Context, tokens *oauth.TokenResponse) (*session.Data, error) {
		email, name, sub, err := fetchUserInfo(ctx, userinfoURL, tokens.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("fetching user info: %w", err)
		}
		if name == "" {
			name = email
		}

		resp, err := identitySvc.ResolveIdentity(ctx, &identitypb.ResolveIdentityRequest{
			Provider:       "oidc",
			ProviderUserId: sub,
		})
		if err == nil {
			return &session.Data{UserID: resp.UserId, Email: email}, nil
		}

		regResp, err := identitySvc.Register(ctx, &identitypb.RegisterRequest{
			Email:          email,
			Name:           name,
			Provider:       "oidc",
			ProviderUserId: sub,
		})
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}
		return &session.Data{UserID: regResp.UserId, Email: email}, nil
	}
}

// fetchUserInfo calls the OIDC userinfo endpoint.
func fetchUserInfo(ctx context.Context, userinfoURL, accessToken string) (email, name, sub string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfoURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", fmt.Errorf("userinfo returned %d: %s", resp.StatusCode, body)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return "", "", "", fmt.Errorf("decoding userinfo response: %w", err)
	}
	return claims.Email, claims.Name, claims.Sub, nil
}
