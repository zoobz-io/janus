import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { GRANTS_WORKSPACE } from "~/constants/grants";
import { acrossApplications } from "~/utils/applications";
import { toPage } from "~/utils/rows";

export const useGrantsPage = () => {
  const api = useAdminApi();

  const grants = useTable("grants", GRANTS_WORKSPACE.slots.grants.widget, {
    fetch: async (params) => {
      const rows = await acrossApplications(api, async (app) => {
        const { grants: list } = await api.applications.grants.list(app.id);
        return list;
      });
      return toPage(rows, params);
    },
  });

  return useWorkspace(GRANTS_WORKSPACE, { grants });
};
