/** Customer organizations and their members. */

import { op } from "../spec.js";

export const tenants = {
  list: op("get", "/tenants"),
  create: op("post", "/tenants"),
  get: op("get", "/tenants/{tenant_id}"),
  update: op("patch", "/tenants/{tenant_id}"),

  members: {
    list: op("get", "/tenants/{tenant_id}/members"),
    add: op("post", "/tenants/{tenant_id}/members"),
    updateRole: op("patch", "/tenants/{tenant_id}/members/{user_id}"),
    remove: op("delete", "/tenants/{tenant_id}/members/{user_id}"),
  },
};
