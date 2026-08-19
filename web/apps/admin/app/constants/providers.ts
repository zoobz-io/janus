import { defineEntity } from "@zoobzio/foundation/definitions/entity";
import { defineWorkspace } from "@zoobzio/foundation/definitions/workspace";

/** The providers endpoint returns bare names; the table rows wrap them. */
export type ProviderRow = {
  provider: string;
};

const providers = defineEntity<ProviderRow>();

export const PROVIDERS_TABLE = providers.defineTable({
  columns: [{ key: "provider", label: "Provider", sortable: true }],
  rowKey: "provider",
});

export const PROVIDERS_WORKSPACE = defineWorkspace({
  columns: 1,
  rows: 1,
  slots: {
    providers: { position: [0, 0], span: [1, 1], widget: PROVIDERS_TABLE },
  },
});
