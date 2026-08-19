import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

export type Application = components["schemas"]["ApplicationResponse"];

const applications = defineEntity<Application>();

export const APPLICATIONS_TABLE = applications.defineTable({
  columns: [
    { key: "name", label: "Name", sortable: true },
    { key: "slug", label: "Slug" },
    { key: "status", label: "Status", sortable: true },
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
