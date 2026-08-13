/**
 * @janus/crest — the crestable provider for janus.
 *
 * Janus is the org's auth service; crestable is how an app consumes one.
 * This package is the bridge between them: `createProvider(schema, options,
 * bridge)` yields a crestable provider whose vendor is the janus public API.
 * The app owns the contract and the bridge; janus owns sessions, licensing,
 * and grants.
 *
 *     import { defineSchema } from "crestable";
 *     import { createProvider } from "@janus/crest";
 *
 *     const provider = createProvider(schema, { api: "/api/janus", app: "janus-admin" }, (session) => ({
 *       id: session.identity.id,
 *       email: session.identity.email,
 *       name: session.identity.display_name,
 *       scopes: session.authorization.tenants[0]?.scopes ?? [],
 *       roles: session.authorization.tenants[0]?.roles ?? [],
 *     }));
 */

export { createProvider, loginUrl } from "./provider.js";
export type {
  Application,
  Authorization,
  AuthorizedTenant,
  Identity,
  Membership,
  Options,
  Session,
} from "./types.js";
