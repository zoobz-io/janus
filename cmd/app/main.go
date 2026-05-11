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
	directorypb "github.com/zoobz-io/aegis/proto/directory/v1"
	entitlementpb "github.com/zoobz-io/aegis/proto/entitlement/v1"
	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/rocco/oauth"
	"github.com/zoobz-io/rocco/session"
	"github.com/zoobz-io/sctx"
	"github.com/zoobz-io/sum"
	"google.golang.org/grpc"

	apicontracts "github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/handlers"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/auth"
	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/internal/mesh"
	"github.com/zoobz-io/janus/internal/observe"
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
	sum.Register[apicontracts.Applications](k, allStores.Applications)
	sum.Register[apicontracts.TenantApplications](k, allStores.TenantApplications)
	sum.Register[apicontracts.UserApplications](k, allStores.UserApplications)

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

	identityServer := mesh.NewIdentityServer(
		allStores.Users, allStores.LinkedIdentities, allStores.Tenants, allStores.Memberships,
		allStores.Applications, allStores.TenantApplications, allStores.UserApplications,
	)

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
		Resolve:     resolveOAuth(issuer, identityServer),
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
	sum.NewBoundary[models.Application](k)
	sum.NewBoundary[models.TenantApplication](k)
	sum.NewBoundary[models.UserApplication](k)
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

	if _, schemaErr := observe.StartSchemaSync(ctx, db, ap); schemaErr != nil {
		return fmt.Errorf("failed to start aperture schema sync: %w", schemaErr)
	}

	// =========================================================================
	// 10. Aegis mesh node with sctx auth
	// =========================================================================

	meshCfg := sum.MustUse[config.Mesh](ctx)

	// Bootstrap sctx admin from file-based keychain.
	keychain := aegis.NewFileKeychain(meshCfg.CertDir)
	admin, err := aegis.NewAdminFromKeychain(ctx, keychain, meshCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to create sctx admin: %w", err)
	}

	// Create guards. We need a self-token to create guards, which requires
	// a temporary assertion against our own cert.
	nodeCert, err := keychain.LoadCertificate(ctx, meshCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to load node certificate: %w", err)
	}
	nodeKey, err := keychain.LoadPrivateKey(ctx, meshCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to load node private key: %w", err)
	}
	nodeAssertion, err := sctx.CreateAssertion(nodeKey, nodeCert)
	if err != nil {
		return fmt.Errorf("failed to create node assertion: %w", err)
	}
	nodeToken, err := admin.Generate(ctx, nodeCert, nodeAssertion)
	if err != nil {
		return fmt.Errorf("failed to generate node token: %w", err)
	}

	// Define guards for each gRPC service method.
	identityResolveGuard, err := admin.CreateGuard(ctx, nodeToken, "identity:resolve")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	identityRegisterGuard, err := admin.CreateGuard(ctx, nodeToken, "identity:register")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	identityReadGuard, err := admin.CreateGuard(ctx, nodeToken, "identity:read")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	sessionManageGuard, err := admin.CreateGuard(ctx, nodeToken, "session:manage")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	directoryReadGuard, err := admin.CreateGuard(ctx, nodeToken, "directory:read")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	directoryWriteGuard, err := admin.CreateGuard(ctx, nodeToken, "directory:write")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}
	entitlementManageGuard, err := admin.CreateGuard(ctx, nodeToken, "entitlement:manage")
	if err != nil {
		return fmt.Errorf("failed to create guard: %w", err)
	}

	// Create gRPC service implementations.
	sessionServer := mesh.NewSessionServer(
		allStores.Sessions, allStores.Users, allStores.Tenants,
		allStores.Applications, allStores.TenantApplications, allStores.UserApplications, allStores.Memberships,
	)
	directoryServer := mesh.NewDirectoryServer(allStores.Users, allStores.Tenants, allStores.Memberships)
	entitlementServer := mesh.NewEntitlementServer(
		allStores.Applications, allStores.TenantApplications, allStores.UserApplications, allStores.Memberships,
	)

	node, err := aegis.NewNodeBuilder().
		WithID(meshCfg.ID).
		WithName(meshCfg.Name).
		WithAddress(meshCfg.Addr()).
		WithCertDir(meshCfg.CertDir).
		WithAdmin(admin).
		WithServices(
			aegis.ServiceInfo{Name: "identity", Version: "v1"},
			aegis.ServiceInfo{Name: "session", Version: "v1"},
			aegis.ServiceInfo{Name: "directory", Version: "v1"},
			aegis.ServiceInfo{Name: "entitlement", Version: "v1"},
		).
		// Identity guards.
		WithGuard(identitypb.IdentityService_ResolveIdentity_FullMethodName, identityResolveGuard).
		WithGuard(identitypb.IdentityService_Register_FullMethodName, identityRegisterGuard).
		WithGuard(identitypb.IdentityService_ListProviders_FullMethodName, identityReadGuard).
		// Session guards.
		WithGuard(sessionpb.SessionService_CreateSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_ValidateSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_RevokeSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_RevokeUserSessions_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_ListUserSessions_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_SubscribeSessionEvents_FullMethodName, sessionManageGuard).
		// Directory guards.
		WithGuard(directorypb.DirectoryService_GetUser_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_GetUserByEmail_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_GetTenant_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_CreateTenant_FullMethodName, directoryWriteGuard).
		WithGuard(directorypb.DirectoryService_UpdateTenant_FullMethodName, directoryWriteGuard).
		// Entitlement guards.
		WithGuard(entitlementpb.EntitlementService_AuthorizeApplication_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_RevokeApplication_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_ListTenantApplications_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_GrantUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_RevokeUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_UpdateUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_ListUserAccess_FullMethodName, entitlementManageGuard).
		// Service registration.
		WithServiceRegistration(func(s *grpc.Server) {
			identitypb.RegisterIdentityServiceServer(s, identityServer)
			sessionpb.RegisterSessionServiceServer(s, sessionServer)
			directorypb.RegisterDirectoryServiceServer(s, directoryServer)
			entitlementpb.RegisterEntitlementServiceServer(s, entitlementServer)
		}).
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

		// Try to resolve existing user.
		resp, err := identitySvc.ResolveIdentity(ctx, &identitypb.ResolveIdentityRequest{
			Provider:       "oidc",
			ProviderUserId: sub,
		})
		if err == nil {
			return &session.Data{
				UserID: resp.UserId,
				Email:  email,
			}, nil
		}

		// User not found — create user + link identity (no tenant yet).
		regResp, err := identitySvc.Register(ctx, &identitypb.RegisterRequest{
			Email:          email,
			Name:           name,
			Provider:       "oidc",
			ProviderUserId: sub,
		})
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}

		return &session.Data{
			UserID: regResp.UserId,
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
