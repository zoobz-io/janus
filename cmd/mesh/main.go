// Package main is the entry point for the janus aegis gRPC mesh node.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zoobz-io/aegis"
	directorypb "github.com/zoobz-io/aegis/proto/directory/v1"
	entitlementpb "github.com/zoobz-io/aegis/proto/entitlement/v1"
	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sctx"
	"github.com/zoobz-io/sum"
	"google.golang.org/grpc"

	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/internal/mesh"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting janus mesh...")
	ctx := context.Background()

	rt, err := boot.Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime: %w", err)
	}
	defer func() { _ = rt.DB.Close() }()
	defer func() { _ = rt.Redis.Close() }()

	if cfgErr := sum.Config[config.Mesh](ctx, rt.K, nil); cfgErr != nil {
		return fmt.Errorf("failed to load mesh config: %w", cfgErr)
	}

	// The mesh servers hold concrete stores directly, so no contract
	// registration is needed — but the registry is frozen so store queries run.
	sum.Freeze(rt.K)
	capitan.Emit(ctx, events.StartupServicesReady)

	// Observability.
	otelProviders, err := boot.OTEL(ctx, "janus-mesh")
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()

	meshCfg := sum.MustUse[config.Mesh](ctx)

	// Bootstrap sctx admin from file-based keychain.
	keychain := aegis.NewFileKeychain(meshCfg.CertDir)
	admin, err := aegis.NewAdminFromKeychain(ctx, keychain, meshCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to create sctx admin: %w", err)
	}

	// A self-token is required to create guards.
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

	guard := func(scope string) (sctx.Guard, error) {
		g, gErr := admin.CreateGuard(ctx, nodeToken, scope)
		if gErr != nil {
			return nil, fmt.Errorf("failed to create guard %q: %w", scope, gErr)
		}
		return g, nil
	}

	identityResolveGuard, err := guard("identity:resolve")
	if err != nil {
		return err
	}
	identityRegisterGuard, err := guard("identity:register")
	if err != nil {
		return err
	}
	identityReadGuard, err := guard("identity:read")
	if err != nil {
		return err
	}
	sessionManageGuard, err := guard("session:manage")
	if err != nil {
		return err
	}
	directoryReadGuard, err := guard("directory:read")
	if err != nil {
		return err
	}
	directoryWriteGuard, err := guard("directory:write")
	if err != nil {
		return err
	}
	entitlementManageGuard, err := guard("entitlement:manage")
	if err != nil {
		return err
	}

	// gRPC service implementations (concrete stores).
	identityServer := mesh.NewIdentityServer(
		rt.Stores.Users, rt.Stores.Accounts, rt.Stores.Tenants, rt.Stores.Memberships,
		rt.Stores.Applications, rt.Stores.Licenses, rt.Stores.Grants,
		rt.Stores.Features, rt.Stores.Scopes,
	)
	sessionServer := mesh.NewSessionServer(
		rt.Stores.Sessions, rt.Stores.Users, rt.Stores.Tenants,
		rt.Stores.Applications, rt.Stores.Licenses, rt.Stores.Grants, rt.Stores.Memberships,
		rt.Stores.Features, rt.Stores.Scopes,
	)
	directoryServer := mesh.NewDirectoryServer(rt.Stores.Users, rt.Stores.Tenants, rt.Stores.Memberships)
	entitlementServer := mesh.NewEntitlementServer(
		rt.Stores.Applications, rt.Stores.Licenses, rt.Stores.Grants, rt.Stores.Memberships,
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
		WithGuard(identitypb.IdentityService_ResolveIdentity_FullMethodName, identityResolveGuard).
		WithGuard(identitypb.IdentityService_Register_FullMethodName, identityRegisterGuard).
		WithGuard(identitypb.IdentityService_ListProviders_FullMethodName, identityReadGuard).
		WithGuard(sessionpb.SessionService_CreateSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_ValidateSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_RevokeSession_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_RevokeUserSessions_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_ListUserSessions_FullMethodName, sessionManageGuard).
		WithGuard(sessionpb.SessionService_SubscribeSessionEvents_FullMethodName, sessionManageGuard).
		WithGuard(directorypb.DirectoryService_GetUser_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_GetUserByEmail_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_GetTenant_FullMethodName, directoryReadGuard).
		WithGuard(directorypb.DirectoryService_CreateTenant_FullMethodName, directoryWriteGuard).
		WithGuard(directorypb.DirectoryService_UpdateTenant_FullMethodName, directoryWriteGuard).
		WithGuard(entitlementpb.EntitlementService_AuthorizeApplication_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_RevokeApplication_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_ListTenantApplications_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_GrantUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_RevokeUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_UpdateUserAccess_FullMethodName, entitlementManageGuard).
		WithGuard(entitlementpb.EntitlementService_ListUserAccess_FullMethodName, entitlementManageGuard).
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
	log.Printf("mesh node listening on %s...", meshCfg.Addr())

	// Block until interrupted.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down janus mesh...")
	return nil
}
