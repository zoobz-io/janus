// Package main is the entry point for the janus admin HTTP API.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco/session"
	"github.com/zoobz-io/sum"

	admincontracts "github.com/zoobz-io/janus/admin/contracts"
	adminhandlers "github.com/zoobz-io/janus/admin/handlers"
	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/auth"
	"github.com/zoobz-io/janus/internal/boot"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting janus admin...")
	ctx := context.Background()

	rt, err := boot.Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime: %w", err)
	}
	defer func() { _ = rt.DB.Close() }()
	defer func() { _ = rt.Redis.Close() }()

	if cfgErr := sum.Config[config.Admin](ctx, rt.K, nil); cfgErr != nil {
		return fmt.Errorf("failed to load admin config: %w", cfgErr)
	}
	if cfgErr := sum.Config[config.Cookie](ctx, rt.K, nil); cfgErr != nil {
		return fmt.Errorf("failed to load cookie config: %w", cfgErr)
	}

	// Admin API contracts — the same shared stores, narrowed to the admin
	// capability boundary.
	sum.Register[admincontracts.Applications](rt.K, rt.Stores.Applications)
	sum.Register[admincontracts.Tenants](rt.K, rt.Stores.Tenants)
	sum.Register[admincontracts.Memberships](rt.K, rt.Stores.Memberships)
	sum.Register[admincontracts.Users](rt.K, rt.Stores.Users)
	sum.Register[admincontracts.Accounts](rt.K, rt.Stores.Accounts)
	sum.Register[admincontracts.Sessions](rt.K, rt.Stores.Sessions)
	sum.Register[admincontracts.Licenses](rt.K, rt.Stores.Licenses)
	sum.Register[admincontracts.Grants](rt.K, rt.Stores.Grants)
	sum.Register[admincontracts.Scopes](rt.K, rt.Stores.Scopes)
	sum.Register[admincontracts.Tiers](rt.K, rt.Stores.Tiers)
	sum.Register[admincontracts.Features](rt.K, rt.Stores.Features)

	sum.Freeze(rt.K)
	capitan.Emit(ctx, events.StartupServicesReady)

	// Observability.
	otelProviders, err := boot.OTEL(ctx, "janus-admin")
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()

	// Authentication: session cookies and bearer session tokens, both validated
	// against the same shared session store as the public API. Admin does not
	// initiate OIDC login — a browser session is established by the public login
	// flow, and its cookie authenticates here because the two APIs share the
	// session store and cookie signing key (a same-origin proxy in front of the
	// admin UI forwards it). Service callers present `Authorization: Bearer <token>`.
	cookieCfg := sum.MustUse[config.Cookie](ctx)
	cookieKey, err := cookieCfg.Key()
	if err != nil {
		return fmt.Errorf("failed to decode cookie sign key: %w", err)
	}
	sessionStore := auth.NewSessionStore(rt.Stores.Sessions, rt.Stores.Users, rt.Redis)
	cookieExtractor := auth.CookieExtractor(sessionStore, session.CookieConfig{SignKey: cookieKey})
	authenticator := auth.NewAuthenticator(rt.Stores.Sessions, rt.Stores.Users, cookieExtractor)
	rt.Svc.Engine().WithAuthenticator(authenticator)

	adminhandlers.ConfigureOpenAPI(rt.Svc.Engine())
	rt.Svc.Handle(adminhandlers.All()...)

	adminCfg := sum.MustUse[config.Admin](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(adminCfg.Port))
	log.Printf("starting admin server on port %d...", adminCfg.Port)
	return rt.Svc.Run("", adminCfg.Port)
}
