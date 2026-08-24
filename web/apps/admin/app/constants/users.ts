import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

export type User = components["schemas"]["UserResponse"];

const users = defineEntity<User>();

// Sortable columns are limited to the search contract's sort allowlist
// (created_at / updated_at) — the fetch action forwards the sort straight
// to POST /users/search.
export const USERS_TABLE = users.defineTable({
  columns: [
    { key: "display_name", label: "Name" },
    { key: "email", label: "Email" },
    { key: "status", label: "Status" },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
    { key: "last_seen_at", label: "Last seen", type: "datetime" },
  ],
  rowKey: "id",
});

export const USERS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    users: { position: [0, 0], span: [1, 1], widget: USERS_TABLE },
  },
});
