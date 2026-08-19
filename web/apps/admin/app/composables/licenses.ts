import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { LICENSES_WORKSPACE } from "~/constants/licenses";
import { acrossApplications } from "~/utils/applications";
import { toPage } from "~/utils/rows";

export const useLicensesPage = () => {
  const api = useAdminApi();

  const licenses = useTable(
    "licenses",
    LICENSES_WORKSPACE.slots.licenses.widget,
    {
      fetch: async (params) => {
        const rows = await acrossApplications(api, async (app) => {
          const { licenses: list } = await api.applications.licenses.list(
            app.id,
          );
          return list;
        });
        return toPage(rows, params);
      },
    },
  );

  return useWorkspace(LICENSES_WORKSPACE, { licenses });
};
