// Package main seeds the development database with representative fake data.
//
// It reuses the shared boot.Init runtime — the same config load, encryption,
// infra connections, store construction and model boundaries every janus binary
// uses — then drives the real stores to create a small but complete mesh:
// applications with scopes/tiers/features, tenants and members, users with
// linked accounts and sessions, licenses, and per-user grants.
//
// The seeder TRUNCATEs the domain tables first so it is safe to re-run. It
// connects to whatever APP_DB_* points at, so it is intended for local dev only
// (defaults target the docker-compose Postgres: localhost:5432/janus).
//
//	make seed        # or: go run ./cmd/seed
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/internal/boot"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	rt, err := boot.Init(ctx)
	if err != nil {
		return fmt.Errorf("initializing runtime: %w", err)
	}
	defer func() { _ = rt.DB.Close() }()
	defer func() { _ = rt.Redis.Close() }()

	// The seeder registers no contracts of its own — it drives the concrete
	// stores directly — but the registry must still be frozen so model
	// boundaries and the encryptor are active.
	sum.Freeze(rt.K)

	if err := reset(ctx, rt.DB); err != nil {
		return fmt.Errorf("resetting tables: %w", err)
	}

	s := &seeder{ctx: ctx, stores: rt.Stores}
	if err := s.seed(); err != nil {
		return err
	}

	log.Printf("seed complete: %d applications, %d tenants, %d users, %d grants, %d sessions",
		s.appCount, len(orgs), len(users), s.grantCount, s.sessionCount)
	return nil
}

// reset clears the domain tables (leaving config intact) so the seeder can be
// run repeatedly. This is the one raw statement in the tool — a dev-only wipe,
// not application data access.
func reset(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `TRUNCATE
		grants, features, licenses, tiers, scopes,
		sessions, accounts, memberships, applications, users, tenants
		RESTART IDENTITY CASCADE`)
	return err
}

// =============================================================================
// Seed data
// =============================================================================

type scopeDef struct {
	name string
	desc string
}

type tierDef struct {
	slug   string
	name   string
	rank   int
	scopes []string // scope names bundled into this tier (features)
}

type appDef struct {
	name   string
	slug   string
	status models.ApplicationStatus
	scopes []scopeDef
	tiers  []tierDef
}

type userDef struct {
	email     string
	name      string
	status    models.UserStatus
	providers []models.ProviderType
}

type grantDef struct {
	app    string // application slug
	roles  []string
	scopes []string
	tier   string // tier slug, or "" for none
}

type memberDef struct {
	email  string // user email
	role   models.UserRole
	grants []grantDef
}

type orgDef struct {
	name     string
	slug     string
	status   models.TenantStatus
	licenses []string // application slugs the tenant is authorized for
	members  []memberDef
}

// Products on the mesh. Each declares its own scope catalog and tiers; tiers
// bundle a subset of scopes as features. Vault is intentionally inactive and
// unused to exercise that state.
var apps = []appDef{
	{
		name: "Nexus", slug: "nexus", status: models.ApplicationStatusActive,
		scopes: []scopeDef{
			{"projects:read", "Read projects"},
			{"projects:write", "Create and edit projects"},
			{"projects:delete", "Delete projects"},
			{"members:manage", "Manage project members"},
			{"billing:read", "View billing"},
		},
		tiers: []tierDef{
			{"free", "Free", 0, []string{"projects:read"}},
			{"pro", "Pro", 1, []string{"projects:read", "projects:write", "members:manage"}},
			{"enterprise", "Enterprise", 2, []string{"projects:read", "projects:write", "projects:delete", "members:manage", "billing:read"}},
		},
	},
	{
		name: "Forge", slug: "forge", status: models.ApplicationStatusActive,
		scopes: []scopeDef{
			{"pipelines:read", "Read pipelines"},
			{"pipelines:run", "Trigger pipeline runs"},
			{"runners:manage", "Manage runners"},
			{"secrets:read", "Read CI secrets"},
		},
		tiers: []tierDef{
			{"starter", "Starter", 0, []string{"pipelines:read"}},
			{"team", "Team", 1, []string{"pipelines:read", "pipelines:run"}},
			{"scale", "Scale", 2, []string{"pipelines:read", "pipelines:run", "runners:manage", "secrets:read"}},
		},
	},
	{
		name: "Atlas", slug: "atlas", status: models.ApplicationStatusActive,
		scopes: []scopeDef{
			{"dashboards:read", "Read dashboards"},
			{"dashboards:write", "Create and edit dashboards"},
			{"data:export", "Export raw data"},
		},
		tiers: []tierDef{
			{"free", "Free", 0, []string{"dashboards:read"}},
			{"pro", "Pro", 1, []string{"dashboards:read", "dashboards:write", "data:export"}},
		},
	},
	{
		name: "Relay", slug: "relay", status: models.ApplicationStatusActive,
		scopes: []scopeDef{
			{"messages:send", "Send messages"},
			{"templates:manage", "Manage templates"},
			{"webhooks:manage", "Manage webhooks"},
		},
		tiers: []tierDef{
			{"free", "Free", 0, []string{"messages:send"}},
			{"growth", "Growth", 1, []string{"messages:send", "templates:manage"}},
			{"enterprise", "Enterprise", 2, []string{"messages:send", "templates:manage", "webhooks:manage"}},
		},
	},
	{
		name: "Vault", slug: "vault", status: models.ApplicationStatusInactive,
		scopes: []scopeDef{
			{"secrets:read", "Read secrets"},
			{"secrets:write", "Write secrets"},
		},
		tiers: []tierDef{
			{"standard", "Standard", 0, []string{"secrets:read", "secrets:write"}},
		},
	},
}

var users = []userDef{
	{"alice@acme.example", "Alice Nguyen", models.UserStatusActive, []models.ProviderType{models.ProviderGitHub, models.ProviderGoogle}},
	{"bob@acme.example", "Bob Martinez", models.UserStatusActive, []models.ProviderType{models.ProviderGitHub}},
	{"carol@acme.example", "Carol Danvers", models.UserStatusActive, nil},
	{"dave@globex.example", "Dave Kim", models.UserStatusActive, []models.ProviderType{models.ProviderGoogle}},
	{"erin@globex.example", "Erin O'Brien", models.UserStatusActive, []models.ProviderType{models.ProviderGitHub}},
	{"frank@initech.example", "Frank Lucas", models.UserStatusActive, []models.ProviderType{models.ProviderAuth0}},
	{"grace@initech.example", "Grace Hopper", models.UserStatusActive, []models.ProviderType{models.ProviderGitHub, models.ProviderGoogle}},
	{"heidi@initech.example", "Heidi Klein", models.UserStatusInactive, nil},
	{"ivan@umbrella.example", "Ivan Petrov", models.UserStatusActive, []models.ProviderType{models.ProviderZitadel}},
	{"judy@umbrella.example", "Judy Chen", models.UserStatusActive, nil},
	{"mallory@hooli.example", "Mallory Singh", models.UserStatusActive, []models.ProviderType{models.ProviderGitHub}},
	{"oscar@hooli.example", "Oscar Reyes", models.UserStatusInactive, []models.ProviderType{models.ProviderGoogle}},
}

// Sessions to create: user email -> the services (issued_by) that minted them.
var sessions = map[string][]string{
	"alice@acme.example":    {"nexus", "forge"},
	"bob@acme.example":      {"nexus"},
	"dave@globex.example":   {"relay"},
	"grace@initech.example": {"nexus", "atlas"},
	"frank@initech.example": {"forge"},
	"mallory@hooli.example": {"atlas"},
}

var orgs = []orgDef{
	{
		name: "Acme Corp", slug: "acme-corp", status: models.TenantStatusActive,
		licenses: []string{"nexus", "forge", "atlas"},
		members: []memberDef{
			{"alice@acme.example", models.UserRoleOwner, []grantDef{
				{"nexus", []string{"admin"}, []string{"projects:read", "projects:write", "members:manage"}, "enterprise"},
				{"forge", []string{"maintainer"}, []string{"pipelines:read", "pipelines:run", "runners:manage"}, "scale"},
				{"atlas", []string{"admin"}, []string{"dashboards:read", "dashboards:write", "data:export"}, "pro"},
			}},
			{"bob@acme.example", models.UserRoleAdmin, []grantDef{
				{"nexus", []string{"editor"}, []string{"projects:read", "projects:write"}, "pro"},
				{"forge", []string{"developer"}, []string{"pipelines:read", "pipelines:run"}, "team"},
			}},
			{"carol@acme.example", models.UserRoleEditor, []grantDef{
				{"nexus", []string{"viewer"}, []string{"projects:read"}, "free"},
			}},
		},
	},
	{
		name: "Globex", slug: "globex", status: models.TenantStatusActive,
		licenses: []string{"nexus", "relay"},
		members: []memberDef{
			{"dave@globex.example", models.UserRoleOwner, []grantDef{
				{"nexus", []string{"admin"}, []string{"projects:read", "projects:write"}, "pro"},
				{"relay", []string{"admin"}, []string{"messages:send", "templates:manage", "webhooks:manage"}, "growth"},
			}},
			{"erin@globex.example", models.UserRoleEditor, []grantDef{
				{"relay", []string{"editor"}, []string{"messages:send"}, "free"},
			}},
		},
	},
	{
		name: "Initech", slug: "initech", status: models.TenantStatusActive,
		licenses: []string{"nexus", "forge", "atlas", "relay"},
		members: []memberDef{
			{"grace@initech.example", models.UserRoleOwner, []grantDef{
				{"nexus", []string{"admin"}, []string{"projects:read", "projects:write", "members:manage"}, "enterprise"},
				{"forge", []string{"maintainer"}, []string{"pipelines:read", "pipelines:run", "runners:manage"}, "scale"},
				{"atlas", []string{"admin"}, []string{"dashboards:read", "dashboards:write", "data:export"}, "pro"},
				{"relay", []string{"admin"}, []string{"messages:send", "templates:manage", "webhooks:manage"}, "enterprise"},
			}},
			{"frank@initech.example", models.UserRoleAdmin, []grantDef{
				{"forge", []string{"developer"}, []string{"pipelines:read", "pipelines:run"}, "team"},
				{"atlas", []string{"viewer"}, []string{"dashboards:read"}, "free"},
			}},
			{"heidi@initech.example", models.UserRoleViewer, []grantDef{
				{"nexus", []string{"viewer"}, []string{"projects:read"}, "free"},
			}},
		},
	},
	{
		name: "Umbrella", slug: "umbrella", status: models.TenantStatusSuspended,
		licenses: []string{"nexus"},
		members: []memberDef{
			{"ivan@umbrella.example", models.UserRoleOwner, []grantDef{
				{"nexus", []string{"admin"}, []string{"projects:read", "projects:write"}, "pro"},
			}},
			{"judy@umbrella.example", models.UserRoleViewer, []grantDef{
				{"nexus", []string{"viewer"}, []string{"projects:read"}, "free"},
			}},
		},
	},
	{
		name: "Hooli", slug: "hooli", status: models.TenantStatusActive,
		licenses: []string{"atlas"},
		members: []memberDef{
			{"mallory@hooli.example", models.UserRoleOwner, []grantDef{
				{"atlas", []string{"admin"}, []string{"dashboards:read", "dashboards:write", "data:export"}, "pro"},
			}},
			{"oscar@hooli.example", models.UserRoleViewer, []grantDef{
				{"atlas", []string{"viewer"}, []string{"dashboards:read"}, "free"},
			}},
		},
	},
}

// =============================================================================
// Seeding
// =============================================================================

type seeder struct {
	ctx    context.Context
	stores *stores.Stores

	appIDs   map[string]string            // slug -> application ID
	scopeIDs map[string]map[string]string // app slug -> scope name -> scope ID
	tierIDs  map[string]map[string]string // app slug -> tier slug -> tier ID
	userIDs  map[string]string            // email -> user ID
	tenants  map[string]string            // slug -> tenant ID

	appCount     int
	grantCount   int
	sessionCount int
}

func (s *seeder) seed() error {
	if err := s.seedApps(); err != nil {
		return err
	}
	if err := s.seedUsers(); err != nil {
		return err
	}
	return s.seedOrgs()
}

func (s *seeder) seedApps() error {
	s.appIDs = map[string]string{}
	s.scopeIDs = map[string]map[string]string{}
	s.tierIDs = map[string]map[string]string{}

	for _, a := range apps {
		app, err := s.stores.Applications.CreateApplication(s.ctx, a.name, a.slug)
		if err != nil {
			return fmt.Errorf("creating application %q: %w", a.slug, err)
		}
		// CreateApplication defaults to active; flip to the desired status.
		if a.status != models.ApplicationStatusActive {
			if _, err := s.stores.Applications.Update(s.ctx, app.ID, a.name, a.status); err != nil {
				return fmt.Errorf("setting status on application %q: %w", a.slug, err)
			}
		}
		s.appIDs[a.slug] = app.ID
		s.appCount++

		s.scopeIDs[a.slug] = map[string]string{}
		for _, sc := range a.scopes {
			scope, err := s.stores.Scopes.Define(s.ctx, app.ID, sc.name, sc.desc)
			if err != nil {
				return fmt.Errorf("defining scope %q on %q: %w", sc.name, a.slug, err)
			}
			s.scopeIDs[a.slug][sc.name] = scope.ID
		}

		s.tierIDs[a.slug] = map[string]string{}
		for _, t := range a.tiers {
			tier, err := s.stores.Tiers.Define(s.ctx, app.ID, t.slug, t.name, t.rank)
			if err != nil {
				return fmt.Errorf("defining tier %q on %q: %w", t.slug, a.slug, err)
			}
			s.tierIDs[a.slug][t.slug] = tier.ID
			for _, scopeName := range t.scopes {
				scopeID, ok := s.scopeIDs[a.slug][scopeName]
				if !ok {
					return fmt.Errorf("tier %q references unknown scope %q on %q", t.slug, scopeName, a.slug)
				}
				if _, err := s.stores.Features.Add(s.ctx, tier.ID, scopeID); err != nil {
					return fmt.Errorf("adding feature %q to tier %q: %w", scopeName, t.slug, err)
				}
			}
		}
	}
	return nil
}

func (s *seeder) seedUsers() error {
	s.userIDs = map[string]string{}
	for _, u := range users {
		user, err := s.stores.Users.CreateUser(s.ctx, u.email, u.name)
		if err != nil {
			return fmt.Errorf("creating user %q: %w", u.email, err)
		}
		if u.status != models.UserStatusActive {
			if _, err := s.stores.Users.Update(s.ctx, user.ID, u.name, u.status); err != nil {
				return fmt.Errorf("setting status on user %q: %w", u.email, err)
			}
		}
		s.userIDs[u.email] = user.ID

		for _, provider := range u.providers {
			subject := fmt.Sprintf("%s|%s", provider, u.email)
			if _, err := s.stores.Accounts.Link(s.ctx, user.ID, provider, subject); err != nil {
				return fmt.Errorf("linking %q account for %q: %w", provider, u.email, err)
			}
		}
	}

	for email, issuers := range sessions {
		userID := s.userIDs[email]
		for _, issuedBy := range issuers {
			if _, _, err := s.stores.Sessions.CreateSession(s.ctx, userID, issuedBy,
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "203.0.113.42"); err != nil {
				return fmt.Errorf("creating session for %q: %w", email, err)
			}
			s.sessionCount++
		}
	}
	return nil
}

func (s *seeder) seedOrgs() error {
	s.tenants = map[string]string{}
	for _, org := range orgs {
		tenant, err := s.stores.Tenants.CreateTenant(s.ctx, org.name, org.slug)
		if err != nil {
			return fmt.Errorf("creating tenant %q: %w", org.slug, err)
		}
		if org.status != models.TenantStatusActive {
			if _, err := s.stores.Tenants.UpdateTenant(s.ctx, tenant.ID, org.name, org.status); err != nil {
				return fmt.Errorf("setting status on tenant %q: %w", org.slug, err)
			}
		}
		s.tenants[org.slug] = tenant.ID

		for _, appSlug := range org.licenses {
			appID, ok := s.appIDs[appSlug]
			if !ok {
				return fmt.Errorf("tenant %q licensed for unknown app %q", org.slug, appSlug)
			}
			if _, err := s.stores.Licenses.Authorize(s.ctx, tenant.ID, appID); err != nil {
				return fmt.Errorf("authorizing %q for %q: %w", appSlug, org.slug, err)
			}
		}

		for _, m := range org.members {
			userID, ok := s.userIDs[m.email]
			if !ok {
				return fmt.Errorf("tenant %q references unknown user %q", org.slug, m.email)
			}
			if _, err := s.stores.Memberships.Create(s.ctx, userID, tenant.ID, m.role); err != nil {
				return fmt.Errorf("adding member %q to %q: %w", m.email, org.slug, err)
			}
			for _, g := range m.grants {
				if err := s.grant(org, tenant.ID, userID, m.email, g); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *seeder) grant(org orgDef, tenantID, userID, email string, g grantDef) error {
	appID, ok := s.appIDs[g.app]
	if !ok {
		return fmt.Errorf("grant for %q references unknown app %q", email, g.app)
	}
	if _, err := s.stores.Grants.Grant(s.ctx, userID, tenantID, appID, g.roles, g.scopes); err != nil {
		return fmt.Errorf("granting %q access to %q in %q: %w", email, g.app, org.slug, err)
	}
	if g.tier != "" {
		tierID, ok := s.tierIDs[g.app][g.tier]
		if !ok {
			return fmt.Errorf("grant for %q references unknown tier %q on %q", email, g.tier, g.app)
		}
		if _, err := s.stores.Grants.SetTier(s.ctx, userID, tenantID, appID, tierID); err != nil {
			return fmt.Errorf("setting tier %q on grant for %q: %w", g.tier, email, err)
		}
	}
	s.grantCount++
	return nil
}
