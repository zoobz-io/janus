import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

/** A scope labeled with the application it belongs to. */
export type ScopeRow = components["schemas"]["ScopeResponse"] & {
  application: string;
};

const scopes = defineEntity<ScopeRow>();

export const SCOPES_TABLE = scopes.defineTable({
  columns: [
    { key: "name", label: "Scope", sortable: true },
    { key: "description", label: "Description" },
    { key: "application", label: "Application", sortable: true },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
  ],
  rowKey: "id",
});

export const SCOPES_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    scopes: { position: [0, 0], span: [1, 1], widget: SCOPES_TABLE },
  },
});
