import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { TENANTS_WORKSPACE } from "~/constants/tenants";

export const useTenantsPage = () => {
  const api = useAdminApi();

  const tenants = useTable("tenants", TENANTS_WORKSPACE.slots.tenants.widget, {
    fetch: async ({ page, pageSize, sortField, sortDirection }) => {
      const { tenants: data, page: meta } = await api.tenants.search({
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
  });

  return useWorkspace(TENANTS_WORKSPACE, { tenants });
};
