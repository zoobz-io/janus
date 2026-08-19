import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { TENANTS_WORKSPACE } from "~/constants/tenants";

export const useTenantsPage = () => {
  const api = useAdminApi();

  const tenants = useTable("tenants", TENANTS_WORKSPACE.slots.tenants.widget, {
    fetch: async ({ page, pageSize }) => {
      const { tenants: data, total } = await api.tenants.list({
        query: {
          limit: String(pageSize),
          offset: String((page - 1) * pageSize),
        },
      });
      return {
        data,
        total,
        pageCount: Math.max(1, Math.ceil(total / pageSize)),
      };
    },
  });

  return useWorkspace(TENANTS_WORKSPACE, { tenants });
};
