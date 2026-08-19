import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

/** A subscription tier labeled with the application it belongs to. */
export type TierRow = components["schemas"]["TierResponse"] & {
  application: string;
};

const tiers = defineEntity<TierRow>();

export const TIERS_TABLE = tiers.defineTable({
  columns: [
    { key: "name", label: "Tier", sortable: true },
    { key: "slug", label: "Slug" },
    { key: "rank", label: "Rank", type: "number", sortable: true },
    { key: "application", label: "Application", sortable: true },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
  ],
  rowKey: "id",
});

export const TIERS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    tiers: { position: [0, 0], span: [1, 1], widget: TIERS_TABLE },
  },
});
