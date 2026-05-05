// Package main is the entry point for the janus auth service.
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/zoobz-io/aegis"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/rocco/oauth"
	"github.com/zoobz-io/rocco/session"
	"github.com/zoobz-io/sum"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"google.golang.org/grpc"

	apicontracts "github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/handlers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/auth"
	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/internal/mesh"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting...")
	ctx := context.Background()

	// =========================================================================
	// 1. Initialize sum service and registry
	// =========================================================================

	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 2. Load configuration
	// =========================================================================

	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load encryption config: %w", err)
	}
	if err := sum.Config[config.OTEL](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load otel config: %w", err)
	}
	if err := sum.Config[config.Mesh](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load mesh config: %w", err)
	}
	if err := sum.Config[config.Auth](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load auth config: %w", err)
	}

	// =========================================================================
	// 3. Encryption setup
	// =========================================================================

	encCfg := sum.MustUse[config.Encryption](ctx)
	encKey, err := hex.DecodeString(encCfg.Key)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}
	aesEncryptor, err := cereal.AES(encKey)
	if err != nil {
		return fmt.Errorf("failed to create AES encryptor: %w", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, aesEncryptor)

	// =========================================================================
	// 4. Connect to infrastructure
	// =========================================================================

	db, err := boot.Database(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	redisClient, err := boot.Redis(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer func() { _ = redisClient.Close() }()
	capitan.Emit(ctx, events.StartupRedisConnected)

	// =========================================================================
	// 5. Create and register stores
	// =========================================================================

	renderer := astqlpg.New()
	allStores := stores.New(db, renderer)

	// Public API contracts.
	sum.Register[apicontracts.Users](k, allStores.Users)
	sum.Register[apicontracts.Sessions](k, allStores.Sessions)
	sum.Register[apicontracts.LinkedIdentities](k, allStores.LinkedIdentities)
	sum.Register[apicontracts.Memberships](k, allStores.Memberships)
	sum.Register[apicontracts.Tenants](k, allStores.Tenants)

	// =========================================================================
	// 6. Authentication (cookie + bearer)
	// =========================================================================

	authCfg := sum.MustUse[config.Auth](ctx)

	// Session store adapter — bridges rocco/session.Store to our sessions table + Redis.
	sessionStore := auth.NewSessionStore(allStores.Sessions, allStores.Users, redisClient)

	// Cookie config for session cookies.
	cookieKey, err := authCfg.CookieKey()
	if err != nil {
		return fmt.Errorf("failed to decode cookie sign key: %w", err)
	}
	cookieCfg := session.CookieConfig{
		SignKey: cookieKey,
	}

	// Cookie-based authenticator (for OAuth login flow).
	cookieExtractor := auth.CookieExtractor(sessionStore, cookieCfg)

	// Dual authenticator: tries cookie, falls back to bearer token.
	authenticator := auth.NewAuthenticator(allStores.Sessions, allStores.Users, cookieExtractor)
	svc.Engine().WithAuthenticator(authenticator)

	// =========================================================================
	// 7. OAuth login/callback/logout/register handlers
	// =========================================================================

	meshServer := mesh.NewServer(allStores.Tenants, allStores.Users, allStores.Memberships, allStores.LinkedIdentities, allStores.Sessions)

	// Build OAuth provider config from OIDC issuer.
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
		Resolve:     resolveOAuth(issuer, meshServer),
	}

	authHandlers, err := handlers.NewAuthHandlers(sessionCfg)
	if err != nil {
		return fmt.Errorf("failed to create auth handlers: %w", err)
	}

	// =========================================================================
	// 8. Register boundaries and freeze
	// =========================================================================

	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Membership](k)
	sum.NewBoundary[models.LinkedIdentity](k)
	sum.NewBoundary[models.Session](k)
	wire.RegisterBoundaries(k)

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 9. Observability
	// =========================================================================

	otelProviders, err := boot.OTEL(ctx, "janus")
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	capitan.Emit(ctx, events.StartupOTELReady)

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// =========================================================================
	// 10. Aegis mesh node
	// =========================================================================

	meshCfg := sum.MustUse[config.Mesh](ctx)

	node, err := aegis.NewNodeBuilder().
		WithID(meshCfg.ID).
		WithName(meshCfg.Name).
		WithAddress(meshCfg.Addr()).
		WithServices(aegis.ServiceInfo{Name: "identity", Version: "v1"}).
		WithServiceRegistration(func(_ *grpc.Server) {
			// gRPC service registration will be added when aegis protos are updated.
		}).
		WithCertDir(meshCfg.CertDir).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build mesh node: %w", err)
	}

	if err := node.StartServer(); err != nil {
		return fmt.Errorf("failed to start mesh server: %w", err)
	}
	defer func() { _ = node.Shutdown() }()
	capitan.Emit(ctx, events.StartupMeshReady)

	// =========================================================================
	// 11. Register handlers and run
	// =========================================================================

	svc.Handle(handlers.AllWithAuth(authHandlers)...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)
	return svc.Run("", appCfg.Port)
}

// resolveOAuth returns the Resolve callback for the session config.
// It looks up the user by linked identity. If not found, creates the user
// and links the identity (but does not create a tenant — the user must
// create one via POST /me/tenants after login).
func resolveOAuth(issuer string, meshServer *mesh.Server) func(context.Context, *oauth.TokenResponse) (*session.Data, error) {
	userinfoURL := issuer + "/oidc/v1/userinfo"

	return func(ctx context.Context, tokens *oauth.TokenResponse) (*session.Data, error) {
		email, name, sub, err := fetchUserInfo(ctx, userinfoURL, tokens.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("fetching user info: %w", err)
		}
		if name == "" {
			name = email
		}

		// Try to resolve existing user.
		resp, err := meshServer.ResolveIdentity(ctx, mesh.ResolveIdentityRequest{
			Provider:        "oidc",
			ExternalSubject: sub,
		})
		if err == nil {
			return &session.Data{
				UserID: resp.UserID,
				Email:  email,
			}, nil
		}

		// User not found — create user + link identity (no tenant yet).
		resp, err = meshServer.Register(ctx, "", "", email, name, "oidc", sub)
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}

		return &session.Data{
			UserID: resp.UserID,
			Email:  email,
		}, nil
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
