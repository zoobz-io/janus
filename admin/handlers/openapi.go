package handlers

import (
	"github.com/zoobz-io/openapi"
	"github.com/zoobz-io/rocco"
)

// ConfigureOpenAPI applies the admin API's OpenAPI metadata to the engine: the
// spec Info, per-tag descriptions, and the domain tag groups. Call it on the
// engine before serving so /openapi and /docs reflect it.
func ConfigureOpenAPI(e *rocco.Engine) {
	e.WithOpenAPIInfo(openapi.Info{
		Title:       "Janus Admin API",
		Description: "Operator API for managing every application, tenant, user, and access grant across the zoobz-io mesh. All endpoints require an authenticated session.",
		Version:     "0.1.0",
	})

	// Application domain — products and who may use them.
	e.WithTag("Applications", "Register and manage the products on the mesh.")
	e.WithTag("Scopes", "The permission catalog each application defines.")
	e.WithTag("Tiers", "Subscription tiers and the scopes they bundle (features).")
	e.WithTag("Licenses", "Tenant authorizations to use an application.")
	e.WithTag("Grants", "A user's access to an application within a tenant.")

	// Directory domain — organizations and people.
	e.WithTag("Tenants", "Customer organizations and their members.")
	e.WithTag("Users", "People, their sessions, and linked identity accounts.")
	e.WithTag("Providers", "Supported external identity providers.")

	e.WithTagGroup("Catalog", "Applications", "Scopes", "Tiers", "Licenses", "Grants")
	e.WithTagGroup("Directory", "Tenants", "Users", "Providers")
}
