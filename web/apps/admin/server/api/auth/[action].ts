/**
 * The auth endpoints the browser transport dials (/api/auth/*). Each
 * request constructs a @janus/authz provider around the h3 event: the
 * caller's cookie is forwarded to the janus public API, and the bridge maps
 * janus's session payload into the admin contract — first entitled tenant
 * wins (operators act through one tenant; janus-ops in practice).
 */

import { defineAuthHandlers } from "@letters-patent/nuxt/server";
import { createProvider } from "@janus/authz";
import { defineSchema } from "letters-patent";
import { getHeader } from "h3";

import { useRuntimeConfig } from "#imports";

import { contract } from "../../../shared/contract";

const schema = defineSchema(contract);

export default defineAuthHandlers(schema, (event) => {
  const cookie = getHeader(event, "cookie");
  return createProvider(
    schema,
    {
      api: useRuntimeConfig(event).janus.authHost,
      app: "janus-admin",
      headers: cookie === undefined ? {} : { cookie },
    },
    (session) => {
      const tenant = session.authorization.tenants[0];
      return {
        id: session.identity.id,
        email: session.identity.email,
        name: session.identity.display_name,
        scopes: tenant?.scopes ?? [],
        roles: tenant?.roles ?? [],
        meta: { tenant: tenant?.tenant_name ?? "" },
      };
    },
  );
});
