/**
 * The admin client factory. Assembles the domain namespaces into one tree and
 * captures it into a Press via `client` from the spec binding. The returned
 * client is fully typed against the admin spec; `createAdminClient` is the
 * only entry point.
 */

import type { paths } from "./schema";

import { definePress } from "openapi-press";

const { op, client } = definePress<paths>();

/**
 * Creates an admin client. With no config it targets the same origin (relative
 * `baseUrl`), which is what the browser wants when a proxy fronts the admin API.
 */
export const createAdminClient = client({
  applications: {
    list: op("get", "/applications"),
    search: op("post", "/applications/search"),
    create: op("post", "/applications"),
    get: op("get", "/applications/{app_id}"),
    update: op("patch", "/applications/{app_id}"),

    grants: {
      list: op("get", "/applications/{app_id}/grants"),
      create: op("post", "/applications/{app_id}/grants"),
      update: op(
        "patch",
        "/applications/{app_id}/grants/{tenant_id}/{user_id}",
      ),
      revoke: op(
        "delete",
        "/applications/{app_id}/grants/{tenant_id}/{user_id}",
      ),
    },

    licenses: {
      list: op("get", "/applications/{app_id}/licenses"),
      authorize: op("post", "/applications/{app_id}/licenses"),
      revoke: op("delete", "/applications/{app_id}/licenses/{tenant_id}"),
    },

    scopes: {
      list: op("get", "/applications/{app_id}/scopes"),
      create: op("post", "/applications/{app_id}/scopes"),
      delete: op("delete", "/applications/{app_id}/scopes/{scope_id}"),
    },

    tiers: {
      list: op("get", "/applications/{app_id}/tiers"),
      create: op("post", "/applications/{app_id}/tiers"),
      delete: op("delete", "/applications/{app_id}/tiers/{tier_id}"),

      features: {
        list: op("get", "/applications/{app_id}/tiers/{tier_id}/features"),
        add: op("post", "/applications/{app_id}/tiers/{tier_id}/features"),
        remove: op(
          "delete",
          "/applications/{app_id}/tiers/{tier_id}/features/{scope_id}",
        ),
      },
    },
  },
  providers: {
    list: op("get", "/providers"),
  },
  tenants: {
    list: op("get", "/tenants"),
    search: op("post", "/tenants/search"),
    create: op("post", "/tenants"),
    get: op("get", "/tenants/{tenant_id}"),
    update: op("patch", "/tenants/{tenant_id}"),

    members: {
      list: op("get", "/tenants/{tenant_id}/members"),
      add: op("post", "/tenants/{tenant_id}/members"),
      updateRole: op("patch", "/tenants/{tenant_id}/members/{user_id}"),
      remove: op("delete", "/tenants/{tenant_id}/members/{user_id}"),
    },
  },
  users: {
    list: op("get", "/users"),
    search: op("post", "/users/search"),
    create: op("post", "/users"),
    get: op("get", "/users/{user_id}"),
    update: op("patch", "/users/{user_id}"),

    accounts: {
      list: op("get", "/users/{user_id}/accounts"),
      unlink: op("delete", "/users/{user_id}/accounts/{account_id}"),
    },

    sessions: {
      list: op("get", "/users/{user_id}/sessions"),
      revokeAll: op("delete", "/users/{user_id}/sessions"),
      revoke: op("delete", "/users/{user_id}/sessions/{session_id}"),
    },
  },
});

/** A live, fully-typed admin client. */
export type AdminClient = ReturnType<typeof createAdminClient>;
