/**
 * The janus vendor domain: the wire shapes the public API answers with and
 * the options a provider instance is constructed from. Bridges consume
 * {@link Session}; everything here is janus's vocabulary, never an app
 * contract's.
 */

/** One tenant membership as reported by `GET /me`. */
export interface Membership {
  tenant_id: string;
  tenant_name: string;
  role: string;
}

/** The caller's identity, as reported by `GET /me`. */
export interface Identity {
  id: string;
  email: string;
  display_name: string;
  memberships: Membership[];
}

/** The application authorization was resolved for. */
export interface Application {
  id: string;
  name: string;
  slug: string;
}

/** One tenant through which the caller holds access to the application. */
export interface AuthorizedTenant {
  tenant_id: string;
  tenant_name: string;
  /** The caller's membership role in the tenant (owner/admin/editor/viewer). */
  role: string;
  /** Application roles on the caller's grant. */
  roles: string[];
  /** Effective scopes: grant scopes unioned with tier features. */
  scopes: string[];
}

/** The caller's resolved authorization, as reported by `GET /me/authorization/{app}`. */
export interface Authorization {
  user_id: string;
  application: Application;
  tenants: AuthorizedTenant[];
}

/**
 * The vendor payload a bridge maps into an app's contract: identity plus the
 * app-scoped authorization, fetched together on every resolve.
 */
export interface Session {
  identity: Identity;
  authorization: Authorization;
}

/** Configuration for a janus provider instance. */
export interface Options {
  /**
   * Janus public API base — an absolute origin (`https://janus.example`) or
   * a same-origin proxy prefix (`/api/janus`). Requests carry the caller's
   * session cookie, so a browser deployment wants the proxy form.
   */
  api: string;
  /** Slug of the application to resolve authorization for (e.g. `janus-admin`). */
  app: string;
  /**
   * Fetch implementation. Defaults to `globalThis.fetch`. Server-side
   * callers (e.g. a Nuxt crest handler) pass one that carries the incoming
   * request's cookie.
   */
  fetch?: typeof globalThis.fetch;
  /** Headers merged into every request (e.g. a forwarded cookie). */
  headers?: HeadersInit;
  /**
   * When true (the default), a valid session with no entitled tenants
   * resolves to signed-out — not entitled to the app means not authenticated
   * for it. Set false to hand such sessions to the bridge anyway (the app
   * then renders its own "no access" state).
   */
  requireEntitlement?: boolean;
}
