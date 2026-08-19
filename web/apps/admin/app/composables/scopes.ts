import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { SCOPES_WORKSPACE } from "~/constants/scopes";
import { acrossApplications } from "~/utils/applications";
import { toPage } from "~/utils/rows";

export const useScopesPage = () => {
  const api = useAdminApi();

  const scopes = useTable("scopes", SCOPES_WORKSPACE.slots.scopes.widget, {
    fetch: async (params) => {
      const rows = await acrossApplications(api, async (app) => {
        const { scopes: list } = await api.applications.scopes.list(app.id);
        return list;
      });
      return toPage(rows, params);
    },
  });

  return useWorkspace(SCOPES_WORKSPACE, { scopes });
};
