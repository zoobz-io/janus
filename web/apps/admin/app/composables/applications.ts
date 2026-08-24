import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { navigateTo } from "#imports";
import { useAdminApi } from "~/composables/api";
import { APPLICATIONS_WORKSPACE } from "~/constants/applications";

export const useApplicationsPage = () => {
  const api = useAdminApi();

  const applications = useTable(
    "applications",
    APPLICATIONS_WORKSPACE.slots.applications.widget,
    {
      fetch: async ({ page, pageSize, sortField, sortDirection }) => {
        const { applications: data, page: meta } =
          await api.applications.search({
            body: {
              page: { number: page, size: pageSize },
              ...(sortField !== null && {
                sort: { field: sortField, order: sortDirection },
              }),
            },
          });
        return {
          data,
          total: meta.total_items,
          pageCount: meta.total_pages,
        };
      },
      actions: {
        view: (row) => navigateTo(`/applications/${row.id}`),
      },
    },
  );

  return useWorkspace(APPLICATIONS_WORKSPACE, { applications });
};
