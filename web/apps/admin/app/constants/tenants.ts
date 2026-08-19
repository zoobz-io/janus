import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

export type Tenant = components["schemas"]["TenantResponse"];

const tenants = defineEntity<Tenant>();

export const TENANTS_TABLE = tenants.defineTable({
  columns: [
    { key: "name", label: "Name" },
    { key: "slug", label: "Slug" },
    { key: "status", label: "Status" },
  ],
  rowKey: "id",
});

export const TENANTS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    tenants: { position: [0, 0], span: [1, 1], widget: TENANTS_TABLE },
  },
});
