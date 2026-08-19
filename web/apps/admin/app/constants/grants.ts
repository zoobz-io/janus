import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

/** A user grant labeled with the application it belongs to. */
export type GrantRow = components["schemas"]["GrantResponse"] & {
  application: string;
};

const grants = defineEntity<GrantRow>();

export const GRANTS_TABLE = grants.defineTable({
  columns: [
    { key: "user_id", label: "User" },
    { key: "tenant_id", label: "Tenant" },
    { key: "application", label: "Application", sortable: true },
    { key: "roles", label: "Roles" },
    { key: "scopes", label: "Scopes" },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
  ],
  rowKey: "id",
});

export const GRANTS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    grants: { position: [0, 0], span: [1, 1], widget: GRANTS_TABLE },
  },
});
