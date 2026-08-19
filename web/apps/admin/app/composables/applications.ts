import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { APPLICATIONS_WORKSPACE } from "~/constants/applications";
import { toPage } from "~/utils/rows";

export const useApplicationsPage = () => {
  const api = useAdminApi();

  const applications = useTable(
    "applications",
    APPLICATIONS_WORKSPACE.slots.applications.widget,
    {
      fetch: async (params) => {
        const { applications: rows } = await api.applications.list();
        return toPage(rows, params);
      },
    },
  );

  return useWorkspace(APPLICATIONS_WORKSPACE, { applications });
};
