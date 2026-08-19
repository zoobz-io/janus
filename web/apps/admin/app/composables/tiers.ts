import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { TIERS_WORKSPACE } from "~/constants/tiers";
import { acrossApplications } from "~/utils/applications";
import { toPage } from "~/utils/rows";

export const useTiersPage = () => {
  const api = useAdminApi();

  const tiers = useTable("tiers", TIERS_WORKSPACE.slots.tiers.widget, {
    fetch: async (params) => {
      const rows = await acrossApplications(api, async (app) => {
        const { tiers: list } = await api.applications.tiers.list(app.id);
        return list;
      });
      return toPage(rows, params);
    },
  });

  return useWorkspace(TIERS_WORKSPACE, { tiers });
};
