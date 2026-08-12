/**
 * Applications and everything scoped beneath one: the grants, licenses, scopes,
 * and tiers (with their features) an application owns. The namespace nesting
 * mirrors the URL hierarchy, so positional args read app-first, left to right.
 */

import { op } from "../spec.js";

export const applications = {
  list: op("get", "/applications"),
  create: op("post", "/applications"),
  get: op("get", "/applications/{app_id}"),
  update: op("patch", "/applications/{app_id}"),

  grants: {
    list: op("get", "/applications/{app_id}/grants"),
    create: op("post", "/applications/{app_id}/grants"),
    update: op("patch", "/applications/{app_id}/grants/{tenant_id}/{user_id}"),
    revoke: op("delete", "/applications/{app_id}/grants/{tenant_id}/{user_id}"),
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
      remove: op("delete", "/applications/{app_id}/tiers/{tier_id}/features/{scope_id}"),
    },
  },
};
