/**
 * The admin portal's crestable contract — its authorization vocabulary,
 * mirroring the janus-admin application's seeded scope catalog and grant
 * roles. Both sides of the wire (the browser transport and the server crest
 * handlers) derive their schema from this one declaration.
 */

import type { Contract } from "crestable";

export const contract = {
  scopes: [
    "directory:read",
    "users:manage",
    "tenants:manage",
    "applications:manage",
  ],
  roles: ["operator", "auditor"],
  meta: {
    /** The tenant the operator is acting through (e.g. "Janus Operations"). */
    tenant: "string",
  },
} as const satisfies Contract;
