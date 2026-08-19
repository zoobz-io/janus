import type { components } from "@janus/admin-sdk";

import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

/** A license labeled with the application it belongs to. */
export type LicenseRow = components["schemas"]["LicenseResponse"] & {
  application: string;
};

const licenses = defineEntity<LicenseRow>();

export const LICENSES_TABLE = licenses.defineTable({
  columns: [
    { key: "application", label: "Application", sortable: true },
    { key: "tenant_id", label: "Tenant" },
    { key: "created_at", label: "Created", type: "datetime", sortable: true },
  ],
  rowKey: "id",
});

export const LICENSES_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    licenses: { position: [0, 0], span: [1, 1], widget: LICENSES_TABLE },
  },
});
