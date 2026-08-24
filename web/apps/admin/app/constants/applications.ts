import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

export type Application = components["schemas"]["ApplicationResponse"];

const applications = defineEntity<Application>();

// Sortable columns are limited to the search contract's sort allowlist
// (created_at / updated_at) — the fetch action forwards the sort straight
// to POST /applications/search.
export const APPLICATIONS_TABLE = applications.defineTable({
  columns: [
    { key: "name", label: "Name" },
    { key: "slug", label: "Slug" },
    { key: "status", label: "Status" },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
    { key: "updated_at", label: "Updated", type: "datetime", sortable: true },
  ],
  rowKey: "id",
});

export const APPLICATIONS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    applications: { position: [0, 0], span: [1, 1], widget: APPLICATIONS_TABLE },
  },
});
