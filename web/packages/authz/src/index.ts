/**
 * @janus/authz — the letters-patent provider for janus.
 *
 * Janus is the org's auth service; letters-patent is how an app consumes
 * one. This package is the bridge between them: `createProvider(schema,
 * options, bridge)` yields a letters-patent provider whose vendor is the
 * janus public API. The app owns the contract and the bridge; janus owns
 * sessions, licensing, and grants.
 *
 *     import { defineSchema } from "letters-patent";
 *     import { createProvider } from "@janus/authz";
 *
 *     const provider = createProvider(schema, { api: "/api/janus", app: "janus-admin" }, (session) => ({
 *       id: session.identity.id,
 *       email: session.identity.email,
 *       name: session.identity.display_name,
 *       scopes: session.authorization.tenants[0]?.scopes ?? [],
 *       roles: session.authorization.tenants[0]?.roles ?? [],
 *     }));
 */

export { createProvider } from "./provider";
export { loginUrl } from "./util";
export type {
  Application,
  Authorization,
  AuthorizedTenant,
  Identity,
  Membership,
  Options,
  Session,
} from "./types";
